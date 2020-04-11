package cmd

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/api"
	gqlgenCfg "github.com/99designs/gqlgen/codegen/config"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
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

			// wd, _ := os.Getwd()
			// watchPaths := []string{
			// 	wd + "/assets",
			// 	wd + "/cmd",
			// 	wd + "/configs",
			// 	wd + "/db",
			// 	wd + "/pkg",
			// 	wd + "/go.sum",
			// 	wd + "/go.mod",
			// 	wd + "/main.go",
			// }
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			signal.Notify(quit, syscall.SIGTERM)

			go func() {
				<-quit
			}()
		},
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
