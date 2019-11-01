package appy

type (
	// App is the core of appy framework which determines how an application is driven.
	App struct {
		cmd       *Cmd
		config    *Config
		logger    *Logger
		server    *Server
		dbManager *DbManager
		support   *Support
	}
)

const (
	// VERSION follows semantic versioning to indicate the framework's release status.
	VERSION = "0.1.0"

	_description = "An opinionated productive web framework that helps scaling business easier."
)

var (
	// Build is the current build type for the application, can be `debug` or `release`. Please take note that this
	// value will be updated to `release` by `go run . build` command.
	Build = "debug"
)

// NewApp initializes App instance that comes with:
//
// cmd - provides appy's built-in commands and allow custom command constructing
// config - provides appy's global configuration
// logger - provides logger
// server - provides the capability to serve HTTP/GRPC requests
// dbManager - manages the databases along with their pool connections
// support - provides utility helpers/extensions
func NewApp() *App {
	support := NewSupport()
	logger := NewLogger(Build)
	config := NewConfig(Build, logger, support)
	server := NewServer()

	cmd := NewCmd()

	if Build == "debug" {
		cmd.AddCommand()
	}

	cmd.AddCommand(
		newSecretCommand(logger),
		newSSLCleanCommand(config, logger),
		newSSLSetupCommand(config, logger, server),
	)

	return &App{
		cmd:     cmd,
		config:  config,
		logger:  logger,
		server:  server,
		support: support,
	}
}

// Cmd returns the app's Cmd instance.
func (a App) Cmd() *Cmd {
	return a.cmd
}

// Config returns the app's Config instance.
func (a App) Config() *Config {
	return a.config
}

// DbManager returns the app's DbManager instance.
func (a App) DbManager() *DbManager {
	return a.dbManager
}

// Logger returns the app's Logger instance.
func (a App) Logger() *Logger {
	return a.logger
}

// Server returns the app's Server instance.
func (a App) Server() *Server {
	return a.server
}

// Support returns the app's Support instance.
func (a App) Support() *Support {
	return a.support
}

// Run starts the application.
func (a App) Run() {
	a.Cmd().Execute()
}
