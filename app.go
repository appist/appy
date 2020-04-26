package appy

import (
	"net/http"

	"github.com/appist/appy/cmd"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/worker"
)

// App is the framework core that drives the application.
type App struct {
	asset     *Asset
	cmd       *Command
	config    *Config
	dbManager *record.Engine
	i18n      *I18n
	logger    *Logger
	mailer    *Mailer
	server    *Server
	worker    *Worker
}

// NewApp initializes an app instance.
func NewApp(assetFS http.FileSystem, appRoot string, viewFuncs map[string]interface{}) *App {
	logger := support.NewLogger()
	asset := support.NewAsset(assetFS, appRoot)
	config := support.NewConfig(asset, logger)
	dbManager := record.NewEngine(logger)
	i18n := support.NewI18n(asset, config, logger)
	ml := mailer.NewEngine(asset, config, i18n, logger, viewFuncs)
	server := pack.NewAppServer(asset, config, i18n, ml, logger, viewFuncs)
	worker := worker.NewEngine(asset, config, dbManager, logger)
	cmd := cmd.NewAppCommand(asset, config, dbManager, logger, server, worker)

	return &App{
		asset,
		cmd,
		config,
		dbManager,
		i18n,
		logger,
		ml,
		server,
		worker,
	}
}

// Command returns the app instance's root command.
func (a *App) Command() *Command {
	return a.cmd
}

// Config returns the app instance's config.
func (a *App) Config() *Config {
	return a.config
}

// DB returns the app instance's specific DB.
func (a *App) DB(name string) DB {
	return a.dbManager.DB(name)
}

// DBManager returns the app instance's DB manager.
func (a *App) DBManager() *DBManager {
	return a.dbManager
}

// I18n returns the app instance's i18n manager.
func (a *App) I18n() *I18n {
	return a.i18n
}

// Logger returns the app instance's logger.
func (a *App) Logger() *Logger {
	return a.logger
}

// Mailer returns the app instance's mailer.
func (a *App) Mailer() *Mailer {
	return a.mailer
}

// Worker returns the app instance's worker.
func (a *App) Worker() *Worker {
	return a.worker
}

// Run starts running the app instance.
func (a *App) Run() error {
	a.server.ServeSPA("/", a.asset.Embedded())
	a.server.ServeNoRoute()

	return a.Command().Execute()
}

// Server returns the app instance's server.
func (a *App) Server() *Server {
	return a.server
}
