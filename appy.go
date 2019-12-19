package appy

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/appist/appy/cmd"
	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
)

type (
	// App is the framework core that drives the application.
	App struct {
		assets  *support.Assets
		command *cmd.Command
		config  *support.Config
		i18n    *support.I18n
		logger  *support.Logger
		server  *ah.Server
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp(static http.FileSystem) *App {
	command := cmd.NewCommand()
	logger := support.NewLogger()
	assets := support.NewAssets(nil, "", static)
	config := support.NewConfig(assets, logger)
	i18n := support.NewI18n(assets, config)
	server := ah.NewServer(assets, config, logger)

	// Setup default middleware
	server.Use(ah.CSRF(config, logger))
	server.Use(ah.RequestID())
	server.Use(ah.RequestLogger(config, logger))
	server.Use(ah.RealIP())
	server.Use(ah.ResponseHeaderFilter())
	server.Use(ah.I18n(i18n))
	server.Use(ah.SessionMngr(config))
	server.Use(ah.HealthCheck(config.HTTPHealthCheckURL))
	server.Use(ah.Prerender(config, logger))
	server.Use(ah.Gzip(config))
	server.Use(ah.Secure(config))
	server.Use(ah.Recovery(logger))

	return &App{
		assets:  assets,
		command: command,
		config:  config,
		i18n:    i18n,
		logger:  logger,
		server:  server,
	}
}

// Cmd returns the cmd instance.
func (a App) Cmd() *cmd.Command {
	return a.command
}

// Config returns the config instance.
func (a App) Config() *support.Config {
	return a.config
}

// I18n returns the I18n instance.
func (a App) I18n() *support.I18n {
	return a.i18n
}

// Logger returns the logger instance.
func (a App) Logger() *support.Logger {
	return a.logger
}

// Server returns the server instance.
func (a App) Server() *ah.Server {
	return a.server
}

// Run starts running the app instance.
func (a App) Run() error {
	return a.command.Execute()
}

// Bootstrap initializes the project layout.
func Bootstrap() {
	_, path, _, _ := runtime.Caller(0)
	appTplPath := filepath.Dir(path) + "/templates/app"

	err := filepath.Walk(appTplPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
