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
		assetsMngr *support.AssetsMngr
		command    *cmd.Command
		config     *support.Config
		logger     *support.Logger
		server     *ah.Server
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
	assetsMngr := support.NewAssetsMngr(nil, "", static)
	config := support.NewConfig(assetsMngr, logger)
	server := ah.NewServer(assetsMngr, config, logger)

	// Setup default middleware
	// server.Use(CSRF(c, l))
	// server.Use(RequestID())
	// server.Use(RequestLogger(c, l))
	server.Use(ah.RealIP())
	// server.Use(ResponseHeaderFilter())
	// server.Use(SessionManager(c))
	// server.Use(HealthCheck(config.HTTPHealthCheckURL))
	// server.Use(Prerender(c, l))
	// server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(ah.Secure(config))
	// server.Use(Recovery(l))

	return &App{
		assetsMngr: assetsMngr,
		command:    command,
		config:     config,
		logger:     logger,
		server:     server,
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
