//+build !test

package appy

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
	"github.com/gorilla/websocket"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/radovskyb/watcher"
	"go.uber.org/zap"
)

type textWithWriteOption struct {
	text    string
	options text.WriteOption
}

var (
	apiServeCmd, webServeCmd, workerCmd             *exec.Cmd
	apiServeCmdReady                                chan bool
	apiServeConsole, webServeConsole, workerConsole *text.Text
	terminalBox                                     *termbox.Terminal
	isGenerating                                                  = false
	watcherPollInterval                             time.Duration = 1
	liveReloadWSConn, liveReloadWSSConn             *websocket.Conn
	colorRe                                         = regexp.MustCompile(`[\x1b|\033]\[[0-9;]*[0-9]+m[^\ .]*[\x1b|\033]\[[0-9;]*[0-9]+m`)
	wordRe                                          = regexp.MustCompile(`[\x1b|\033]\[[0-9;]*[0-9]+m(.*)[\x1b|\033]\[[0-9;]*[0-9]+m`)
)

func newStartCommand(logger *Logger, server *Server) *Command {
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

			wd, _ := os.Getwd()
			watchPaths := []string{
				wd + "/assets",
				wd + "/cmd",
				wd + "/configs",
				wd + "/db",
				wd + "/pkg",
				wd + "/go.sum",
				wd + "/go.mod",
				wd + "/main.go",
			}
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			signal.Notify(quit, syscall.SIGTERM)

			go func() {
				<-quit
				killAllCommands()
			}()

			if _, err := os.Stat(wd + "/package.json"); !os.IsNotExist(err) {
				go runWebServeCmd(logger, server)
			}

			go runAPIServeCmd(logger)
			go runWorkerCmd(logger)
			go runLiveReloadServer(logger, server)

			go func() {
				var err error
				terminalBox, err = termbox.New()
				if err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}
				defer terminalBox.Close()

				apiServeConsole, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
				if err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}

				webServeConsole, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
				if err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}

				workerConsole, err = text.New(text.RollContent(), text.WrapAtRunes(), text.WrapAtWords())
				if err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}

				ctx, cancel := context.WithCancel(context.Background())
				ctn, err := container.New(
					terminalBox,
					container.SplitVertical(
						container.Left(
							container.SplitHorizontal(
								container.Top(
									container.Border(linestyle.Light),
									container.BorderTitle(" Backend "),
									container.PlaceWidget(apiServeConsole),
								),
								container.Bottom(
									container.Border(linestyle.Light),
									container.BorderTitle(" Frontend (webpack-dev-server) "),
									container.PlaceWidget(webServeConsole),
								),
							),
						),
						container.Right(
							container.Border(linestyle.Light),
							container.BorderTitle(" Worker "),
							container.PlaceWidget(workerConsole),
						),
					),
				)

				if err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}

				tQuit := func(k *terminalapi.Keyboard) {
					if k.Key == -26 {
						cancel()
						time.Sleep(250 * time.Millisecond)
						killAllCommands()
						os.Exit(0)
					}
				}

				if err := termdash.Run(ctx, terminalBox, ctn, termdash.KeyboardSubscriber(tQuit)); err != nil {
					quit <- os.Kill
					killAllCommands()
					logger.Fatal(err)
				}
			}()

			watch(logger, watchPaths, func(e watcher.Event) {
				watchHandler(e, logger)
			})
		},
	}
}

func watchHandler(e watcher.Event, logger *Logger) {
	if isGenerating {
		return
	}

	isGenerating = true
	if strings.Contains(e.Path, ".gql") || strings.Contains(e.Path, ".graphql") || strings.Contains(e.Path, "pkg/graphql/config.yml") {
		apiServeConsole.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
		apiServeConsole.Write("INFO", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		apiServeConsole.Write(" * Generating GraphQL boilerplate code...\n")

		err := generateGQL()
		if err != nil {
			apiServeConsole.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
			apiServeConsole.Write("ERROR", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			apiServeConsole.Write(" " + err.Error() + "\n")
		}

		isGenerating = false
		return
	}

	gqlgenConfig, _ := gqlgenLoadConfig()
	if gqlgenConfig != nil && (strings.Contains(e.Path, gqlgenConfig.Model.Filename) || (strings.Contains(e.Path, gqlgenConfig.Exec.Filename) && e.Op == watcher.Remove)) {
		isGenerating = false
		return
	}

	isGenerating = false
	go runAPIServeCmd(logger)
	go runWorkerCmd(logger)
}

func gqlgenLoadConfig() (*gqlgenCfg.Config, error) {
	wd, _ := os.Getwd()
	return gqlgenCfg.LoadConfig(wd + "/pkg/graphql/config.yml")
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

func killAPIServeCmd() {
	if apiServeCmd != nil {
		syscall.Kill(-apiServeCmd.Process.Pid, syscall.SIGINT)
		apiServeCmd = nil
	}
}

func runAPIServeCmd(logger *Logger) {
	killAPIServeCmd()
	time.Sleep(500 * time.Millisecond)

	apiServeCmd = exec.Command("go", "run", ".", "serve")
	apiServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	apiServeCmdOut, _ := apiServeCmd.StdoutPipe()
	apiServeCmdErr, _ := apiServeCmd.StderrPipe()

	apiServeCmdReady = make(chan bool, 1)
	go func() {
		<-apiServeCmdReady
		time.Sleep(500 * time.Millisecond)

		if liveReloadWSConn != nil {
			liveReloadWSConn.WriteMessage(websocket.TextMessage, []byte("reload"))
		}

		if liveReloadWSSConn != nil {
			liveReloadWSSConn.WriteMessage(websocket.TextMessage, []byte("reload"))
		}
	}()

	go func(stdout io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		out := bufio.NewScanner(stdout)
		for out.Scan() {
			results := preprocessText(out.Text())

			for _, result := range results {
				if result.text == "" {
					continue
				}

				if err := apiServeConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}

			if len(results) > 0 && strings.Contains(results[len(results)-1].text, "* Listening on") {
				apiServeCmdReady <- true
			}
		}
	}(apiServeCmdOut)

	go func(stderr io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		err := bufio.NewScanner(stderr)
		for err.Scan() {
			results := preprocessText(err.Text())

			for _, result := range results {
				if result.text == "" {
					continue
				}

				if err := apiServeConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}

			if len(results) > 0 && strings.Contains(results[len(results)-1].text, "* Listening on") {
				apiServeCmdReady <- true
			}
		}
	}(apiServeCmdErr)

	apiServeConsole.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
	apiServeConsole.Write("INFO", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
	apiServeConsole.Write(" * Compiling...\n")
	apiServeCmd.Run()
}

func killWorkerCmd() {
	if workerCmd != nil {
		syscall.Kill(-workerCmd.Process.Pid, syscall.SIGINT)
		workerCmd = nil
	}
}

func runWorkerCmd(logger *Logger) {
	killWorkerCmd()
	time.Sleep(500 * time.Millisecond)

	workerCmd = exec.Command("go", "run", ".", "worker")
	workerCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	workerCmdOut, _ := workerCmd.StdoutPipe()
	workerCmdErr, _ := workerCmd.StderrPipe()

	go func(stdout io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		out := bufio.NewScanner(stdout)
		for out.Scan() {
			results := preprocessText(out.Text())

			for _, result := range results {
				if result.text == "" || string(result.text) == "\n" {
					continue
				}

				if err := workerConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}
		}
	}(workerCmdOut)

	go func(stderr io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		err := bufio.NewScanner(stderr)
		for err.Scan() {
			results := preprocessText(err.Text())

			for _, result := range results {
				if result.text == "" || string(result.text) == "\n" {
					continue
				}

				if err := workerConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}
		}
	}(workerCmdErr)

	workerConsole.Write(time.Now().Format("2006-01-02T15:04:05.000-0700") + " ")
	workerConsole.Write("INFO", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
	workerConsole.Write(" * Compiling...\n")
	workerCmd.Run()
}

func killWebServeCmd() {
	if webServeCmd != nil {
		syscall.Kill(-webServeCmd.Process.Pid, syscall.SIGINT)
		webServeCmd = nil
	}
}

func runWebServeCmd(logger *Logger, server *Server) {
	killWebServeCmd()
	wd, _ := os.Getwd()
	ssrPaths := []string{}
	for _, route := range server.Routes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	webServeCmd = exec.Command("npm", "start")
	webServeCmd.Dir = wd
	webServeCmd.Env = os.Environ()
	webServeCmd.Env = append(webServeCmd.Env, "APPY_SSR_ROUTES="+strings.Join(ssrPaths, ","))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_HOST="+server.Config().HTTPHost)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_PORT="+server.Config().HTTPPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_PORT="+server.Config().HTTPSSLPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_ENABLED="+strconv.FormatBool(server.Config().HTTPSSLEnabled))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_CERT_PATH="+server.Config().HTTPSSLCertPath)
	webServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	webServeCmdOut, _ := webServeCmd.StdoutPipe()
	webServeCmdErr, _ := webServeCmd.StderrPipe()

	go func(stdout io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		out := bufio.NewScanner(stdout)
		for out.Scan() {
			results := preprocessText(out.Text())

			for _, result := range results {
				if result.text == "" {
					continue
				}

				if err := webServeConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}
		}
	}(webServeCmdOut)

	go func(stderr io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		err := bufio.NewScanner(stderr)
		for err.Scan() {
			results := preprocessText(err.Text())

			for _, result := range results {
				if result.text == "" {
					continue
				}

				if err := webServeConsole.Write(result.text, result.options); err != nil {
					killAllCommands()
					logger.Fatal(err)
				}
			}
		}
	}(webServeCmdErr)

	webServeCmd.Run()
}

func killAllCommands() {
	killWebServeCmd()
	killAPIServeCmd()
	killWorkerCmd()
}

func convertText(input string) textWithWriteOption {
	colorWordPairs := wordRe.FindStringSubmatch(input)

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
	matchIndexes := colorRe.FindAllSubmatchIndex([]byte(output), -1)
	results := []textWithWriteOption{
		textWithWriteOption{
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
			results = append(results, convertText(output[ptr-1:len(output)]))
		}
	}

	if len(results) > 0 {
		results[len(results)-1].text += "\n"
	}

	return results
}

func runLiveReloadServer(logger *Logger, server *Server) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	wsHandler := http.NewServeMux()
	wsHandler.HandleFunc(LiveReloadPath, func(w http.ResponseWriter, r *http.Request) {
		var err error

		liveReloadWSConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			killAllCommands()
			logger.Fatal(err)
		}

		for {
			_, _, err := liveReloadWSConn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := &http.Server{
		Addr:    server.Config().HTTPHost + ":" + LiveReloadWSPort,
		Handler: wsHandler,
	}
	ws.ErrorLog = zap.NewStdLog(logger.Desugar())

	wssHandler := http.NewServeMux()
	wssHandler.HandleFunc(LiveReloadPath, func(w http.ResponseWriter, r *http.Request) {
		var err error

		liveReloadWSSConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			killAllCommands()
			logger.Fatal(err)
		}

		for {
			_, _, err := liveReloadWSSConn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	wss := &http.Server{
		Addr:    server.Config().HTTPHost + ":" + LiveReloadWSSPort,
		Handler: wssHandler,
	}
	wss.ErrorLog = zap.NewStdLog(logger.Desugar())

	go func() {
		if server.Config().HTTPSSLEnabled {
			err := wss.ListenAndServeTLS(server.Config().HTTPSSLCertPath+"/cert.pem", server.Config().HTTPSSLCertPath+"/key.pem")
			if err != http.ErrServerClosed {
				killAllCommands()
				logger.Fatal(err)
			}
		}
	}()

	err := ws.ListenAndServe()
	if err != http.ErrServerClosed {
		killAllCommands()
		logger.Fatal(err)
	}
}

func watch(logger *Logger, watchPaths []string, callback func(e watcher.Event)) {
	w := watcher.New()
	defer w.Close()

	w.SetMaxEvents(2)

	r := regexp.MustCompile(`.(development|env|go|gql|graphql|ini|json|html|production|test|toml|txt|yml)$`)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				killAllCommands()
				logger.Fatal(r)
			}
		}()

		for {
			select {
			case event := <-w.Event:
				callback(event)
			case err := <-w.Error:
				killAllCommands()
				logger.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	for _, watchPath := range watchPaths {
		w.AddRecursive(watchPath)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	go func() {
		<-quit
		w.Close()
	}()

	if err := w.Start(time.Second * watcherPollInterval); err != nil {
		killAllCommands()
		logger.Fatal(err)
	}
}
