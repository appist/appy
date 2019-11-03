package appy

import (
	"bufio"
	"bytes"
	"io"
	"log"
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
	"github.com/radovskyb/watcher"
)

var (
	gqlgenConfig        *gqlgenCfg.Config
	apiServeCmd         *exec.Cmd
	webServeCmd         *exec.Cmd
	webServeCmdReady    chan os.Signal
	isGenerating                      = false
	watcherPollInterval time.Duration = 1
)

func newStartCommand(s *Server) *Cmd {
	return &Cmd{
		Use:   "start",
		Short: "Runs the GRPC/HTTP web server in development watch mode (debug build only)",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(s.config, s.logger)

			if s.config.HTTPSSLEnabled == true && !s.IsSSLCertExisted() {
				s.logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			wd, _ := os.Getwd()
			watchPaths := []string{
				wd + "/cmd",
				wd + "/db",
				wd + "/pkg",
				wd + "/go.sum",
				wd + "/go.mod",
				wd + "/main.go",
			}
			quit := make(chan os.Signal, 1)
			webServeCmdReady = make(chan os.Signal, 1)

			signal.Notify(quit, os.Interrupt)
			signal.Notify(quit, syscall.SIGTERM)

			go func() {
				<-quit
				killWebServeCmd()
				killAPIServeCmd()
			}()

			if _, err := os.Stat(wd + "/package.json"); !os.IsNotExist(err) {
				go runWebServeCmd(s)
			}

			go func() {
				<-webServeCmdReady
				runAPIServeCmd(s)
			}()

			watch(s, watchPaths, func(e watcher.Event) {
				watchHandler(e, s)
			})
		},
	}
}

func watchHandler(e watcher.Event, s *Server) {
	if isGenerating == true {
		return
	}

	isGenerating = true
	if strings.Contains(e.Path, ".gql") || strings.Contains(e.Path, ".graphql") {
		s.logger.Info("* Generating GraphQL boilerplate code...")

		err := generateGQL(s)
		if err != nil {
			s.logger.Info(err.Error())
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
	go runAPIServeCmd(s)
}

func gqlgenLoadConfig() (*gqlgenCfg.Config, error) {
	wd, _ := os.Getwd()
	return gqlgenCfg.LoadConfig(wd + "/pkg/graphql/config.yml")
}

func generateGQL(s *Server) error {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	gqlgenConfig, _ := gqlgenLoadConfig()
	if err := api.Generate(gqlgenConfig); err != nil {
		return err
	}

	return nil
}

func killAPIServeCmd() {
	if apiServeCmd != nil {
		syscall.Kill(-apiServeCmd.Process.Pid, syscall.SIGINT)
		apiServeCmd = nil
	}
}

func runAPIServeCmd(s *Server) {
	killAPIServeCmd()
	time.Sleep(500 * time.Millisecond)
	apiServeCmd = exec.Command("go", "run", ".", "serve")
	apiServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	apiServeCmd.Stdout = os.Stdout
	apiServeCmd.Stderr = os.Stderr
	s.logger.Info("* Compiling...")
	apiServeCmd.Run()
}

func killWebServeCmd() {
	if webServeCmd != nil {
		syscall.Kill(-webServeCmd.Process.Pid, syscall.SIGINT)
		webServeCmd = nil
	}
}

func runWebServeCmd(s *Server) {
	wd, _ := os.Getwd()
	ssrPaths := []string{}
	for _, route := range s.Routes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	webServeCmd = exec.Command("npm", "run", "start")
	webServeCmd.Dir = wd
	webServeCmd.Env = os.Environ()
	webServeCmd.Env = append(webServeCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_HOST="+s.config.HTTPHost)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_PORT="+s.config.HTTPPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_PORT="+s.config.HTTPSSLPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_ENABLED="+strconv.FormatBool(s.config.HTTPSSLEnabled))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_CERT_PATH="+s.config.HTTPSSLCertPath)
	webServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	webServeCmdOut, _ := webServeCmd.StdoutPipe()
	webServeCmdErr, _ := webServeCmd.StderrPipe()

	go func(stdout io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				s.logger.Fatal(r)
			}
		}()

		timeRe := regexp.MustCompile(` [0-9]+ms`)
		isFirstTime := true
		isWDSCompiling := false
		out := bufio.NewScanner(stdout)

		for out.Scan() {
			outText := strings.Trim(out.Text(), " ")

			if outText == "" && (isWDSCompiling || isFirstTime) {
				continue
			}

			if strings.Contains(outText, "｢wdm｣") || strings.HasPrefix(outText, "> ") || (isWDSCompiling && strings.Contains(outText, "｢wds｣")) || strings.HasPrefix(outText, "error") {
				continue
			}

			if strings.Contains(outText, "Compiling...") || strings.Contains(outText, "｢wds｣") {
				isWDSCompiling = true
				s.logger.Info("* [wds] Compiling...")
			} else if strings.Contains(outText, "Compiled successfully in") {
				isWDSCompiling = false
				s.logger.Infof("* [wds] Compiled successfully in%s", timeRe.FindStringSubmatch(outText)[0])

				if isFirstTime {
					isFirstTime = false
					close(webServeCmdReady)
				}
			} else if strings.HasPrefix(outText, "ERROR  Failed to compile") {
				s.logger.Info("* [wds] Failed to compile.")
				s.logger.Info("")
			} else {
				if len(outText) > 0 {
					s.logger.Info(outText)
				}
			}
		}
	}(webServeCmdOut)

	go func(stderr io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				s.logger.Fatal(r)
			}
		}()

		err := bufio.NewScanner(stderr)
		fatalErr := ""
		for err.Scan() {
			fatalErr = fatalErr + strings.Trim(err.Text(), " ") + "\n\t"
		}

		killAPIServeCmd()
		time.Sleep(1 * time.Second)

		if fatalErr != "" {
			s.logger.Fatal(fatalErr)
		}
	}(webServeCmdErr)

	webServeCmd.Run()
}

func watch(s *Server, watchPaths []string, callback func(e watcher.Event)) {
	w := watcher.New()
	defer w.Close()

	w.SetMaxEvents(2)

	r := regexp.MustCompile(`.(development|env|go|gql|graphql|ini|json|html|production|test|toml|yml)$`)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				s.logger.Fatal(r)
			}
		}()

		for {
			select {
			case event := <-w.Event:
				callback(event)
			case err := <-w.Error:
				s.logger.Fatal(err)
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
		s.logger.Fatal(err)
	}
}
