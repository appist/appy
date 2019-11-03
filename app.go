package appy

import (
	"html/template"
	"net/http"
	"os"
)

type (
	// App is the core of appy framework which determines how an application is driven.
	App struct {
		cmd       *Cmd
		config    *Config
		logger    *Logger
		server    *Server
		dbManager *DbManager
	}
)

const (
	// DebugBuild tends to be slow as it includes debug lvl logging which is more verbose.
	DebugBuild = "debug"

	// ReleaseBuild tends to be faster as it excludes debug lvl logging.
	ReleaseBuild = "release"

	// VERSION follows semantic versioning to indicate the framework's release status.
	VERSION = "0.1.0"

	_description = "An opinionated productive web framework that helps scaling business easier."
)

var (
	// Build is the current build type for the application, can be `debug` or `release`. Please take note that this
	// value will be updated to `release` by `go run . build` command.
	Build = DebugBuild
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes App instance that comes with:
//
// cmd - provides appy's built-in commands and allow custom command constructing
// config - provides appy's global configuration
// logger - provides logger
// server - provides the capability to serve HTTP/GRPC requests
// dbManager - manages the databases along with their pool connections
func NewApp(assets http.FileSystem, viewHelper template.FuncMap) *App {
	cmd := NewCmd()
	logger := NewLogger(Build)
	config := NewConfig(Build, logger, assets)
	dbManager := NewDbManager(logger)
	server := NewServer(config, logger, assets, viewHelper)
	server.InitSSR()

	if Build == DebugBuild {
		cmd.AddCommand(
			newBuildCommand(server),
			newDbSchemaDumpCommand(config, dbManager, logger),
			newGMigrationCommand(config, dbManager, logger),
			newStartCommand(server),
		)
	}

	cmd.AddCommand(
		newConfigDecryptCommand(config, logger),
		newConfigEncryptCommand(config, logger),
		newDbCreateCommand(config, dbManager, logger),
		newDbDropCommand(config, dbManager, logger),
		newDbMigrateCommand(config, dbManager, logger),
		newDbMigrateStatusCommand(config, dbManager, logger),
		newDbRollbackCommand(config, dbManager, logger),
		newDbSchemaLoadCommand(config, dbManager, logger),
		newDcUpCommand(logger, assets),
		newDcDownCommand(logger, assets),
		newMiddlewareCommand(config, logger, server),
		newRoutesCommand(config, logger, server),
		newSecretCommand(logger),
		newServeCommand(dbManager, server),
		newSSLCleanCommand(logger, server),
		newSSLSetupCommand(logger, server),
	)

	return &App{
		cmd:       cmd,
		config:    config,
		dbManager: dbManager,
		logger:    logger,
		server:    server,
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

// Run starts the application.
func (a App) Run() {
	// Must be located right before the server runs due to CSR utilizes `NoRoute` to achieve pretty URL navigation
	// with HTML5 history API.
	a.server.InitCSR()

	// Start executing the root command.
	a.Cmd().Execute()
}
