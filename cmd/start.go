package cmd

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/api"
	gqlgenCfg "github.com/99designs/gqlgen/codegen/config"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
	"github.com/gorilla/websocket"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/radovskyb/watcher"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"
)

var (
	terminalReadyNotifier chan bool
	quitNotifier          chan os.Signal
	watcherPollInterval   time.Duration = 1
	watcherRegex                        = regexp.MustCompile(`.(development|env|go|gql|graphql|ini|json|html|production|test|toml|txt|yml)$`)
	colorRegex                          = regexp.MustCompile(`[\x1b|\033]\[[0-9;]*[0-9]+m[^\ .]*[\x1b|\033]\[[0-9;]*[0-9]+m`)
	wordRegex                           = regexp.MustCompile(`[\x1b|\033]\[[0-9;]*[0-9]+m(.*)[\x1b|\033]\[[0-9;]*[0-9]+m`)
)

func newStartCommand(logger *support.Logger, server *pack.Server) *Command {
	return &Command{
		Use:   "start",
		Short: "Run the HTTP/HTTPS web server with `webpack-dev-server` in development watch mode (only available in debug build)",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
			}

			if server.Config().HTTPSSLEnabled && !server.IsSSLCertExisted() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `ssl:setup` first.")
			}

			quitNotifier = make(chan os.Signal, 1)
			signal.Notify(quitNotifier, os.Interrupt)
			signal.Notify(quitNotifier, syscall.SIGINT, syscall.SIGTERM)

			term := &terminal{}
			go func(t *terminal) {
				<-quitNotifier
				killProcesses(t)
				time.Sleep(500 * time.Millisecond)
				os.Exit(0)
			}(term)

			terminalReadyNotifier = make(chan bool, 1)
			go initTerminal(term)
			<-terminalReadyNotifier

			go initLiveReloadServer(logger, server, term)
			go execServeCmd(server, term)
			go execWebCmd(server, term)
			go execWorkCmd(term)
			initWatcher(server, term)
		},
	}
}

func execServeCmd(server *pack.Server, term *terminal) {
	err := killProcess(term.serveCmd)
	if err != nil {
		term.error(term.serve, err.Error())
		return
	}

	term.serveCmd = exec.Command("go", "run", ".", "serve")
	outPipe, _ := term.serveCmd.StdoutPipe()
	errPipe, _ := term.serveCmd.StderrPipe()

	go term.streamPipe(term.serve, outPipe, true)
	go term.streamPipe(term.serve, errPipe, true)

	term.info(term.serve, "* Compiling...")
	if err := term.serveCmd.Start(); err != nil {
		term.error(term.serve, err.Error())
	}
}

func execWebCmd(server *pack.Server, term *terminal) {
	wd, _ := os.Getwd()
	if _, err := os.Stat(wd + "/package.json"); os.IsNotExist(err) {
		return
	}

	if err := killProcess(term.webCmd); err != nil {
		term.error(term.web, err.Error())
		return
	}

	ssrPaths := []string{}
	for _, route := range server.Routes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	term.webCmd = exec.Command("npm", "start")
	term.webCmd.Dir = wd
	term.webCmd.Env = os.Environ()
	term.webCmd.Env = append(term.webCmd.Env, "APPY_SSR_ROUTES="+strings.Join(ssrPaths, ","))
	term.webCmd.Env = append(term.webCmd.Env, "HTTP_HOST="+server.Config().HTTPHost)
	term.webCmd.Env = append(term.webCmd.Env, "HTTP_PORT="+server.Config().HTTPPort)
	term.webCmd.Env = append(term.webCmd.Env, "HTTP_SSL_PORT="+server.Config().HTTPSSLPort)
	term.webCmd.Env = append(term.webCmd.Env, "HTTP_SSL_ENABLED="+strconv.FormatBool(server.Config().HTTPSSLEnabled))
	term.webCmd.Env = append(term.webCmd.Env, "HTTP_SSL_CERT_PATH="+server.Config().HTTPSSLCertPath)
	outPipe, _ := term.webCmd.StdoutPipe()
	errPipe, _ := term.webCmd.StderrPipe()

	go term.streamPipe(term.web, outPipe, false)
	go term.streamPipe(term.web, errPipe, false)

	if err := term.webCmd.Start(); err != nil {
		term.error(term.web, err.Error())
	}
}

func execWorkCmd(term *terminal) {
	err := killProcess(term.workCmd)
	if err != nil {
		term.error(term.work, err.Error())
		return
	}

	term.workCmd = exec.Command("go", "run", ".", "work")
	outPipe, _ := term.workCmd.StdoutPipe()
	errPipe, _ := term.workCmd.StderrPipe()

	go term.streamPipe(term.work, outPipe, true)
	go term.streamPipe(term.work, errPipe, true)

	term.info(term.work, "* Compiling...")
	if err := term.workCmd.Start(); err != nil {
		term.error(term.work, err.Error())
	}
}

func initLiveReloadServer(logger *support.Logger, server *pack.Server, term *terminal) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	wsHandler := http.NewServeMux()
	wsHandler.HandleFunc(pack.LiveReloadPath, func(w http.ResponseWriter, r *http.Request) {
		var err error

		term.lrWsConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			term.error(term.serve, err.Error())
		}

		for {
			_, _, err := term.lrWsConn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := &http.Server{
		Addr:    server.Config().HTTPHost + ":" + pack.LiveReloadWSPort,
		Handler: wsHandler,
	}
	ws.ErrorLog = zap.NewStdLog(logger.Desugar())

	wssHandler := http.NewServeMux()
	wssHandler.HandleFunc(pack.LiveReloadPath, func(w http.ResponseWriter, r *http.Request) {
		var err error

		term.lrWssConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			term.error(term.serve, err.Error())
		}

		for {
			_, _, err := term.lrWssConn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	wss := &http.Server{
		Addr:    server.Config().HTTPHost + ":" + pack.LiveReloadWSSPort,
		Handler: wssHandler,
	}
	wss.ErrorLog = zap.NewStdLog(logger.Desugar())

	go func() {
		if server.Config().HTTPSSLEnabled {
			err := wss.ListenAndServeTLS(server.Config().HTTPSSLCertPath+"/cert.pem", server.Config().HTTPSSLCertPath+"/key.pem")
			if err != http.ErrServerClosed {
				term.error(term.serve, err.Error())
			}
		}
	}()

	err := ws.ListenAndServe()
	if err != http.ErrServerClosed {
		term.error(term.serve, err.Error())
	}
}

func initTerminal(term *terminal) {
	var err error

	term.box, err = termbox.New()
	if err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}
	defer term.box.Close()

	term.serve, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
	if err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}

	term.web, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
	if err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}

	term.work, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
	if err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctn, err := container.New(
		term.box,
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle(" Backend "),
						container.PlaceWidget(term.serve),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" Frontend (webpack-dev-server) "),
						container.PlaceWidget(term.web),
					),
				),
			),
			container.Right(
				container.Border(linestyle.Light),
				container.BorderTitle(" Worker "),
				container.PlaceWidget(term.work),
			),
		),
	)

	if err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}

	terminalReadyNotifier <- true

	quitHandler := func(k *terminalapi.Keyboard) {
		if k.Key == -26 {
			cancel()
			quitNotifier <- os.Interrupt
		}
	}

	if err := termdash.Run(ctx, term.box, ctn, termdash.KeyboardSubscriber(quitHandler)); err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}
}

func initWatcher(server *pack.Server, term *terminal) {
	wd, _ := os.Getwd()
	paths := []string{
		wd + "/assets",
		wd + "/cmd",
		wd + "/configs",
		wd + "/db",
		wd + "/pkg",
		wd + "/go.sum",
		wd + "/go.mod",
		wd + "/main.go",
	}

	w := watcher.New()
	defer w.Close()

	w.SetMaxEvents(1)
	w.AddFilterHook(watcher.RegexFilterHook(watcherRegex, false))

	for _, p := range paths {
		w.AddRecursive(p)
	}

	go func(t *terminal) {
		defer func(tt *terminal) {
			if err := recover(); err != nil {
				tt.err = err.(error)
				quitNotifier <- os.Interrupt
			}
		}(t)

		for {
			select {
			case event := <-w.Event:
				watcherHandler(event, server, term)
			case err := <-w.Error:
				t.err = err
				quitNotifier <- os.Interrupt
			case <-w.Closed:
				return
			}
		}
	}(term)

	if err := w.Start(watcherPollInterval * time.Second); err != nil {
		term.err = err
		quitNotifier <- os.Interrupt
	}
}

func generateGQL() error {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	gqlgenConfig, _ := gqlgenLoadConfig()
	return api.Generate(gqlgenConfig)
}

func gqlgenLoadConfig() (*gqlgenCfg.Config, error) {
	wd, _ := os.Getwd()

	return gqlgenCfg.LoadConfig(wd + "/pkg/graphql/config.yml")
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}

	p, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return err
	}

	children, err := p.Children()
	if err != nil {
		return err
	}

	for _, child := range children {
		err = child.Terminate()
		if err != nil {
			return err
		}
	}

	return p.Terminate()
}

func killProcesses(term *terminal) {
	err := killProcess(term.serveCmd)
	if err != nil {
		term.error(term.serve, err.Error())
	}

	err = killProcess(term.webCmd)
	if err != nil {
		term.error(term.web, err.Error())
	}

	err = killProcess(term.workCmd)
	if err != nil {
		term.error(term.work, err.Error())
	}

	if term.err != nil {
		term.error(term.serve, term.err.Error())
		term.error(term.web, term.err.Error())
		term.error(term.work, term.err.Error())
	}
}

func watcherHandler(e watcher.Event, server *pack.Server, term *terminal) {
	if term.isGeneratingGQL {
		return
	}

	term.isGeneratingGQL = true
	if strings.Contains(e.Path, ".gql") || strings.Contains(e.Path, ".graphql") || strings.Contains(e.Path, "pkg/graphql/config.yml") {
		term.info(term.serve, "* Generating GraphQL boilerplate code...")

		err := generateGQL()
		if err != nil {
			term.error(term.serve, " "+err.Error())
		}

		term.isGeneratingGQL = false
		return
	}

	gqlgenConfig, _ := gqlgenLoadConfig()
	if gqlgenConfig != nil && (strings.Contains(e.Path, gqlgenConfig.Model.Filename) || (strings.Contains(e.Path, gqlgenConfig.Exec.Filename) && e.Op == watcher.Remove)) {
		term.isGeneratingGQL = false
		return
	}

	term.isGeneratingGQL = false

	if !term.isCompiling {
		term.isCompiling = true
		go execServeCmd(server, term)
		go execWorkCmd(term)
	}
}

type terminal struct {
	box                          *termbox.Terminal
	isCompiling, isGeneratingGQL bool
	lrWsConn, lrWssConn          *websocket.Conn
	serve, web, work             *text.Text
	serveCmd, webCmd, workCmd    *exec.Cmd
	err                          error
}

func (t *terminal) error(container *text.Text, msg string) {
	container.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
	container.Write("ERROR", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
	container.Write(" " + msg + "\n")
}

func (t *terminal) info(container *text.Text, msg string) {
	container.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
	container.Write("INFO", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
	container.Write(" " + msg + "\n")
}

func (t *terminal) warn(container *text.Text, msg string) {
	container.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
	container.Write("WARN", text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	container.Write(" " + msg + "\n")
}

func (t *terminal) streamPipe(container *text.Text, reader io.ReadCloser, skipLineBreak bool) {
	defer func(tt *terminal) {
		if err := recover(); err != nil {
			tt.err = err.(error)
			quitNotifier <- os.Interrupt
		}
	}(t)

	out := bufio.NewScanner(reader)
	for out.Scan() {
		results := preprocessText(out.Text())

		for _, result := range results {
			if result.text == "" || (skipLineBreak && string(result.text) == "\n") {
				continue
			}

			if err := container.Write(result.text, result.options); err != nil {
				t.err = err
				quitNotifier <- os.Interrupt
			}
		}

		if len(results) > 0 && strings.Contains(results[len(results)-1].text, "* Listening on") {
			t.isCompiling = false
			time.Sleep(500 * time.Millisecond)

			if t.lrWsConn != nil {
				t.lrWsConn.WriteMessage(websocket.TextMessage, []byte("reload"))
			}

			if t.lrWssConn != nil {
				t.lrWssConn.WriteMessage(websocket.TextMessage, []byte("reload"))
			}
		}
	}
}

type textWithWriteOption struct {
	text    string
	options text.WriteOption
}

func convertText(input string) textWithWriteOption {
	colorWordPairs := wordRegex.FindStringSubmatch(input)

	if len(colorWordPairs) != 2 {
		return textWithWriteOption{
			text:    input,
			options: text.WriteCellOpts(),
		}
	}

	splits := strings.Split(input, colorWordPairs[1])

	var options text.WriteOption
	switch splits[0] {
	case "\x1b[30m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorBlack))
	case "\x1b[31m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
	case "\x1b[32m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorGreen))
	case "\x1b[33m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
	case "\x1b[34m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorBlue))
	case "\x1b[35m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))
	case "\x1b[36m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorCyan))
	case "\x1b[40m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorBlack))
	case "\x1b[41m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRed))
	case "\x1b[42m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorGreen))
	case "\x1b[43m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorYellow))
	case "\x1b[44m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorBlue))
	case "\x1b[45m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorMagenta))
	case "\x1b[47m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorWhite))
	case "\x1b[90m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(85, 85, 85)))
	case "\x1b[91m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(255, 85, 85)))
	case "\x1b[92m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(85, 255, 85)))
	case "\x1b[93m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(255, 255, 85)))
	case "\x1b[94m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(85, 85, 255)))
	case "\x1b[95m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(255, 85, 255)))
	case "\x1b[96m":
		options = text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(85, 255, 255)))
	case "\x1b[100m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(85, 85, 85)))
	case "\x1b[101m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(255, 85, 85)))
	case "\x1b[102m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(85, 255, 85)))
	case "\x1b[103m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(255, 255, 85)))
	case "\x1b[104m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(85, 85, 255)))
	case "\x1b[105m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(255, 85, 255)))
	case "\x1b[106m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorRGB24(85, 255, 255)))
	case "\x1b[107m":
		options = text.WriteCellOpts(cell.BgColor(cell.ColorWhite))
	default:
		options = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
	}

	return textWithWriteOption{
		text:    colorWordPairs[1],
		options: options,
	}
}

func preprocessText(input string) []textWithWriteOption {
	output := strings.ReplaceAll(input, "\t", " ")
	output = strings.Trim(output, "\n")
	matchIndexes := colorRegex.FindAllSubmatchIndex([]byte(output), -1)
	results := []textWithWriteOption{
		{
			text:    output,
			options: text.WriteCellOpts(),
		},
	}

	if len(matchIndexes) > 0 {
		results = []textWithWriteOption{}
		ptr := 0
		currMatchIdx := 0

		for currMatchIdx < len(matchIndexes) {
			idxStart := matchIndexes[currMatchIdx][0]
			idxEnd := matchIndexes[currMatchIdx][1]

			if len(results) > 0 {
				lastResult := results[len(results)-1]
				lastChar := lastResult.text[len(lastResult.text)-1]

				if string(lastChar) != " " {
					results = append(results, convertText(" "))
				}
			}

			if ptr < idxStart {
				results = append(results, convertText(output[ptr:idxStart]))
				ptr = idxStart
			} else if ptr >= idxStart && ptr <= idxEnd {
				results = append(results, convertText(output[idxStart:idxEnd]))
				ptr = idxEnd + 1
				currMatchIdx++
			}
		}

		if ptr < len(output) {
			results = append(results, convertText(output[ptr-1:]))
		}
	}

	if len(results) > 0 {
		results[len(results)-1].text += "\n"
	}

	return results
}
