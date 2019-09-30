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

	ah "github.com/appist/appy/http"

	"github.com/99designs/gqlgen/api"
	gqlgenCfg "github.com/99designs/gqlgen/codegen/config"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var (
	gqlgenConfig        *gqlgenCfg.Config
	httpServeCmd        *exec.Cmd
	webServeCmd         *exec.Cmd
	isGenerating                      = false
	watcherPollInterval time.Duration = 1
)

// NewHTTPDevCommand runs the HTTP/HTTPS web server in development watch mode, only available for debug build.
func NewHTTPDevCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "http:dev",
		Short: "Run the HTTP/HTTPS web server in development watch mode, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			checkSSLCerts(s)
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
				killHTTPServeCmd()
				killWebServeCmd()
			}()

			go runHTTPServeCmd()
			if _, err := os.Stat(wd + "/web"); !os.IsNotExist(err) {
				time.Sleep(3 * time.Second)
				go runWebServeCmd(s)
			}

			watchFileChanges(watchPaths, fileChangesHandler)
		},
	}
}

func fileChangesHandler(e watcher.Event) {
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
	go runHTTPServeCmd()
}

func gqlgenLoadConfig() (*gqlgenCfg.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return gqlgenCfg.LoadConfig(wd + "/app/graphql/config.yml")
}

func generateGQL() {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	logger.Info("Generating GraphQL boilerplate code...")
	gqlgenConfig, _ := gqlgenLoadConfig()
	if err := api.Generate(gqlgenConfig); err != nil {
		logger.Error(err)
	}
}

func killHTTPServeCmd() {
	if httpServeCmd != nil {
		syscall.Kill(-httpServeCmd.Process.Pid, syscall.SIGINT)
		httpServeCmd = nil
	}
}

func runHTTPServeCmd() {
	killHTTPServeCmd()
	httpServeCmd = exec.Command("go", "run", ".", "http:serve")
	httpServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	httpServeCmd.Stdout = os.Stdout
	httpServeCmd.Stderr = os.Stderr
	httpServeCmd.Run()
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
	for _, route := range s.GetAllRoutes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	webServeCmd = exec.Command("npm", "run", "serve")
	webServeCmd.Dir = wd + "/web"
	webServeCmd.Env = os.Environ()
	webServeCmd.Env = append(webServeCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
	webServeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	webServeCmdOut, _ := webServeCmd.StdoutPipe()
	webServeCmdErr, _ := webServeCmd.StderrPipe()

	go func(stdout io.ReadCloser, stderr io.ReadCloser) {
		firstTime := true
		scheme := "http"
		port, _ := strconv.Atoi(s.Config.HTTPPort)
		if s.Config.HTTPSSLEnabled == true {
			scheme = "https"
			port, _ = strconv.Atoi(s.Config.HTTPSSLPort)
		}

		hosts := getIPHosts(s)
		host := fmt.Sprintf("%s://%s:%s", scheme, hosts[0], strconv.Itoa(port+1))
		timeRe := regexp.MustCompile(` [0-9]+ms`)

		for {
			out := bufio.NewScanner(stdout)
			for out.Scan() {
				t := out.Text()

				if strings.Contains(t, "Starting development server") {
					logger.Info("* [wds] Starting...")
				} else if strings.Contains(t, "Compiling...") {
					logger.Info("* [wds] Compiling...")
				} else if strings.Contains(t, "Compiled successfully in") {
					logger.Infof("* [wds] Compiled successfully in%s", timeRe.FindStringSubmatch(t)[0])

					if firstTime == true {
						firstTime = false
						logger.Infof("* [wds] Listening on %s", host)
					}
				}
			}
		}
	}(webServeCmdOut, webServeCmdErr)

	webServeCmd.Run()
}

func watchFileChanges(watchPaths []string, callback func(e watcher.Event)) {
	w := watcher.New()
	defer w.Close()

	w.SetMaxEvents(2)

	r := regexp.MustCompile(`.(development|env|go|gql|graphql|ini|json|html|production|test|toml|yml)$`)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				killHTTPServeCmd()
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
