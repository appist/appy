package appy

import (
	"html/template"
	"net/http"
	"os"

	appycmd "github.com/appist/appy/internal/cmd"
	appyhttp "github.com/appist/appy/internal/http"
	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

type (
	// App is the core of appy framework which determines how an application is driven.
	App struct {
		cmd       *Command
		config    *Config
		dbManager *DbManager
		logger    *Logger
		server    *Server
	}

	// Command defines what a command line can do.
	Command = appycmd.Command

	// Config defines the application settings.
	Config = appysupport.Config

	// DbManager manages multiple databases.
	DbManager = appyorm.DbManager

	// Logger provides the logging functionality.
	Logger = appysupport.Logger

	// Server is the engine that serves HTTP requests.
	Server = appyhttp.Server
)

var (
	app *App
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// Init initializes App instance that comes with:
//
// cmd - provides appy's built-in commands and allow custom command constructing
// config - provides appy's global configuration
// logger - provides logger
// server - provides the capability to serve HTTP requests
// dbManager - manages the databases along with their pool connections
func Init(assets http.FileSystem, viewHelper template.FuncMap) {
	cmd := appycmd.NewCommand()
	logger := appysupport.NewLogger()
	config := appysupport.NewConfig(logger, assets)
	dbManager := appyorm.NewDbManager(logger)
	server := appyhttp.NewServer(config, logger, assets, viewHelper)
	server.InitSSR()

	if appysupport.Build == appysupport.DebugBuild {
		cmd.AddCommand()
	}

	cmd.AddCommand()

	app = &App{
		cmd:       cmd,
		config:    config,
		dbManager: dbManager,
		logger:    logger,
		server:    server,
	}
}

// Cmd returns the app's Cmd instance.
func (a App) Cmd() *appycmd.Command {
	return a.cmd
}

// Config returns the app's Config instance.
func (a App) Config() *appysupport.Config {
	return a.config
}

// DbManager returns the app's DbManager instance.
func (a App) DbManager() *appyorm.DbManager {
	return a.dbManager
}

// Logger returns the app's Logger instance.
func (a App) Logger() *appysupport.Logger {
	return a.logger
}

// Server returns the app's Server instance.
func (a App) Server() *appyhttp.Server {
	return a.server
}

// Run starts the application.
func (a App) Run() error {
	// Must be located right before the server runs due to CSR utilizes `NoRoute` to achieve pretty URL navigation
	// with HTML5 history API.
	// a.server.InitCSR()

	// Start executing the root command.
	return a.Cmd().Execute()
}

// SetPlugins initializes the plugins.
func (a *App) SetPlugins(cb func(*App)) {
	cb(a)
}

// Default returns the app instance that is attached to appy module.
func Default() *App {
	return app
}
