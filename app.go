package appy

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
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

	// TestSuite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
	TestSuite = suite.Suite
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

	// CreateTestContext returns a fresh router w/ context for testing purposes.
	CreateTestContext = gin.CreateTestContext

	// RunTestSuite takes a testing suite and runs all of the tests attached to it.
	RunTestSuite = suite.Run

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
		newDbSeedCommand(config, dbManager, logger),
		newDcDownCommand(logger, assets),
		newDcUpCommand(logger, assets),
		newDcRestartCommand(logger, assets),
		newMiddlewareCommand(config, logger, server),
		newRoutesCommand(config, logger, server),
		newSecretCommand(logger),
		newServeCommand(dbManager, server),
		newSSLCleanCommand(logger, server),
		newSSLSetupCommand(logger, server),
	)

	app = &App{
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
func (a App) Run() error {
	// Must be located right before the server runs due to CSR utilizes `NoRoute` to achieve pretty URL navigation
	// with HTML5 history API.
	a.server.InitCSR()

	// Start executing the root command.
	return a.Cmd().Execute()
}

// Default returns the app instance that is attached to appy module.
func Default() *App {
	return app
}
