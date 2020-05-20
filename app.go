package appy

import (
	"net/http"
	"os"

	"github.com/appist/appy/cmd"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/worker"
)

// App is the framework core that drives the application.
type App struct {
	asset     *support.Asset
	cmd       *cmd.Command
	config    *support.Config
	dbManager *record.Engine
	i18n      *support.I18n
	logger    *support.Logger
	mailer    *mailer.Engine
	server    *pack.Server
	worker    *worker.Engine
}

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
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
func (a *App) Command() *cmd.Command {
	return a.cmd
}

// Config returns the app instance's config.
func (a *App) Config() *support.Config {
	return a.config
}

// DB returns the app instance's specific DB.
func (a *App) DB(name string) record.DBer {
	return a.dbManager.DB(name)
}

// DBManager returns the app instance's DB manager.
func (a *App) DBManager() *record.Engine {
	return a.dbManager
}

// I18n returns the app instance's i18n manager.
func (a *App) I18n() *support.I18n {
	return a.i18n
}

// Logger returns the app instance's logger.
func (a *App) Logger() *support.Logger {
	return a.logger
}

// Mailer returns the app instance's mailer.
func (a *App) Mailer() *mailer.Engine {
	return a.mailer
}

// Model returns the layer that represents business data and logic.
func (a *App) Model(m interface{}, opts ...record.ModelOption) record.Modeler {
	return record.NewModel(a.dbManager, m, opts...)
}

// Worker returns the app instance's worker.
func (a *App) Worker() *worker.Engine {
	return a.worker
}

// Run starts running the app instance.
func (a *App) Run() error {
	a.server.ServeSPA("/", a.asset.Embedded())
	a.server.ServeNoRoute()

	return a.Command().Execute()
}

// Server returns the app instance's server.
func (a *App) Server() *pack.Server {
	return a.server
}
