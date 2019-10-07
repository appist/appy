package cmd

import (
	"bufio"
	"bytes"
	"fmt"
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
	ah "github.com/appist/appy/http"
	"github.com/radovskyb/watcher"

	"github.com/spf13/cobra"
)

var (
	gqlgenConfig        *gqlgenCfg.Config
	apiServeCmd         *exec.Cmd
	webServeCmd         *exec.Cmd
	isGenerating                      = false
	watcherPollInterval time.Duration = 1
)

// NewStartCommand runs the GRPC/HTTP web server in development watch mode, only available for debug build.
func NewStartCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Run the GRPC/HTTP web server in development watch mode, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			if s.Config.HTTPSSLEnabled == true && !s.IsSSLCertsExist() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			wd, _ := os.Getwd()
			watchPaths := []string{
				wd + "/app",
				wd + "/cmd",
				wd + "/go.sum",
				wd + "/go.mod",
				wd + "/main.go",
			}
			quit := make(chan os.Signal, 1)

			signal.Notify(quit, os.Interrupt)
			signal.Notify(quit, syscall.SIGTERM)

			go func() {
				<-quit
				killAPIServeCmd()
				killWebServeCmd()
			}()

			go runAPIServeCmd()
			if _, err := os.Stat(wd + "/" + ah.CSRRoot + "/package.json"); !os.IsNotExist(err) {
				time.Sleep(3 * time.Second)
				go runWebServeCmd(s)
			}
			watch(watchPaths, watchHandler)
		},
	}
}

func watchHandler(e watcher.Event) {
	if isGenerating == true {
		return
	}

	isGenerating = true
	if strings.Contains(e.Path, ".gql") || strings.Contains(e.Path, ".graphql") {
		generateGQL()
		isGenerating = false
		return
	}

	gqlgenConfig, _ := gqlgenLoadConfig()
	if gqlgenConfig != nil && (strings.Contains(e.Path, gqlgenConfig.Model.Filename) || (strings.Contains(e.Path, gqlgenConfig.Exec.Filename) && e.Op == watcher.Remove)) {
		isGenerating = false
		return
	}

	isGenerating = false
	go runAPIServeCmd()
}

func gqlgenLoadConfig() (*gqlgenCfg.Config, error) {
	wd, _ := os.Getwd()
	return gqlgenCfg.LoadConfig(wd + "/app/graphql/config.yml")
}

func generateGQL() {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	logger.Info("* Generating GraphQL boilerplate code...")
	gqlgenConfig, _ := gqlgenLoadConfig()
	if err := api.Generate(gqlgenConfig); err != nil {
		logger.Error(err)
	}
}

func killAPIServeCmd() {
	if apiServeCmd != nil {
		syscall.Kill(-apiServeCmd.Process.Pid, syscall.SIGINT)
		apiServeCmd = nil
	}
}

func runAPIServeCmd() {
	killAPIServeCmd()
	time.Sleep(500 * time.Millisecond)
	apiServeCmd = exec.Command("go", "run", ".", "serve")
	apiServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	apiServeCmd.Stdout = os.Stdout
	apiServeCmd.Stderr = os.Stderr
	logger.Info("* Compiling...")
	apiServeCmd.Run()
}

func killWebServeCmd() {
	if webServeCmd != nil {
		syscall.Kill(-webServeCmd.Process.Pid, syscall.SIGINT)
		webServeCmd = nil
	}
}

func runWebServeCmd(s *ah.ServerT) {
	wd, _ := os.Getwd()
	ssrPaths := []string{}
	for _, route := range s.Routes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	webServeCmd = exec.Command("npm", "run", "serve")
	webServeCmd.Dir = wd + "/" + ah.CSRRoot
	webServeCmd.Env = os.Environ()
	webServeCmd.Env = append(webServeCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_HOST="+s.Config.HTTPHost)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_PORT="+s.Config.HTTPPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_PORT="+s.Config.HTTPSSLPort)
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_ENABLED="+strconv.FormatBool(s.Config.HTTPSSLEnabled))
	webServeCmd.Env = append(webServeCmd.Env, "HTTP_SSL_CERT_PATH="+s.Config.HTTPSSLCertPath)
	webServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	webServeCmdOut, _ := webServeCmd.StdoutPipe()
	webServeCmdErr, _ := webServeCmd.StderrPipe()

	go func(stdout io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				logger.Fatal(r)
			}
		}()

		scheme := "http"
		port, _ := strconv.Atoi(s.Config.HTTPPort)
		if s.Config.HTTPSSLEnabled == true {
			scheme = "https"
			port, _ = strconv.Atoi(s.Config.HTTPSSLPort)
		}

		hosts, _ := s.Hosts()
		host := fmt.Sprintf("%s://%s:%s", scheme, hosts[0], strconv.Itoa(port+1))
		timeRe := regexp.MustCompile(` [0-9]+ms`)
		stdoutBlank := true
		out := bufio.NewScanner(stdout)

		for out.Scan() {
			outText := strings.Trim(out.Text(), " ")

			if outText == "" {
				if stdoutBlank {
					continue
				}

				stdoutBlank = true
			} else {
				stdoutBlank = false
			}

			if strings.Contains(outText, "｢wds｣") || strings.Contains(outText, "Use Ctrl+C to close it") || strings.Contains(outText, "> ") ||
				strings.Contains(outText, "App running at") || strings.Contains(outText, "- Local:") || strings.Contains(outText, "Using ") ||
				strings.Contains(outText, "Webpack Bundle Analyzer is started at") || strings.Contains(outText, "Starting type checking service...") ||
				strings.Contains(outText, "Type checking in progress...") {
				continue
			}

			if strings.Contains(outText, "Starting development server") {
				logger.Info("Starting webpack development server...")
			} else if strings.Contains(outText, "- Network:") {
				logger.Infof("App is running at: %s", host)
				logger.Infof("Webpack Bundle Analyzer is running at: http://%s:8888", hosts[0])
			} else if strings.Contains(outText, "Compiling...") {
				logger.Info("Compiling...")
			} else if strings.Contains(outText, "Compiled successfully in") {
				logger.Infof("Compiled successfully in%s", timeRe.FindStringSubmatch(outText)[0])
			} else {
				logger.Info(outText)
			}
		}
	}(webServeCmdOut)

	go func(stderr io.ReadCloser) {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				logger.Fatal(r)
			}
		}()

		err := bufio.NewScanner(stderr)
		fatalErr := ""
		for err.Scan() {
			errText := strings.Trim(err.Text(), " ")

			if strings.Contains(errText, "npm ERR!") {
				break
			}

			fatalErr = fatalErr + errText + "\n\t"
		}

		killAPIServeCmd()
		time.Sleep(1 * time.Second)
		logger.Fatal(fatalErr)
	}(webServeCmdErr)

	webServeCmd.Run()
}

func watch(watchPaths []string, callback func(e watcher.Event)) {
	w := watcher.New()
	defer w.Close()

	w.SetMaxEvents(2)

	r := regexp.MustCompile(`.(development|env|go|gql|graphql|ini|json|html|production|test|toml|yml)$`)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				killAPIServeCmd()
				logger.Fatal(r)
			}
		}()

		for {
			select {
			case event := <-w.Event:
				callback(event)
			case err := <-w.Error:
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
		logger.Fatal(err)
	}
}
