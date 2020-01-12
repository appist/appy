//+build !test

package appy

import (
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
		asset      *Asset
		command    *Command
		config     *Config
		i18n       *I18n
		logger     *Logger
		mailer     *Mailer
		server     *Server
		support    Supporter
		viewEngine *ViewEngine
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp(asset *Asset, viewFuncs map[string]interface{}) *App {
	support := &Support{}
	logger := NewLogger()
	config := NewConfig(asset, logger, support)
	i18n := NewI18n(asset, config, logger)
	viewEngine := NewViewEngine(asset, config, logger)
	server := NewServer(asset, config, logger, support)
	mailer := NewMailer(asset, config, i18n, logger, server, viewFuncs)

	// Setup the default middleware.
	server.Use(AttachLogger(logger))
	server.Use(AttachI18n(i18n))
	server.Use(AttachMailer(mailer))
	server.Use(AttachViewEngine(asset, config, logger, viewFuncs))
	server.Use(RealIP())
	server.Use(RequestID())
	server.Use(RequestLogger(config, logger))
	server.Use(Gzip(config))
	server.Use(HealthCheck(config.HTTPHealthCheckURL))
	server.Use(Prerender(config, logger))
	server.Use(CSRF(config, logger, support))
	server.Use(Secure(config))
	server.Use(APIOnlyResponse())
	server.Use(SessionManager(config))
	server.Use(Recovery(logger))

	command := NewRootCommand()
	command.AddCommand(newRoutesCommand(config, logger, server))
	command.AddCommand(newSecretCommand(logger))
	command.AddCommand(newSSLSetupCommand(logger, server))
	command.AddCommand(newSSLTeardownCommand(logger, server))

	return &App{
		asset:      asset,
		command:    command,
		config:     config,
		i18n:       i18n,
		logger:     logger,
		mailer:     mailer,
		server:     server,
		support:    support,
		viewEngine: viewEngine,
	}
}

// Asset returns the app instance's asset.
func (a *App) Asset() *Asset {
	return a.asset
}

// Command returns the app instance's root command.
func (a *App) Command() *Command {
	return a.command
}

// Config returns the app instance's config.
func (a *App) Config() *Config {
	return a.config
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

// Server returns the app instance's server.
func (a *App) Server() *Server {
	return a.server
}

// Support returns the app instance's support.
func (a *App) Support() Supporter {
	return a.support
}

// ViewEngine returns the app instance's view engine.
func (a *App) ViewEngine() *ViewEngine {
	return a.viewEngine
}

// Run starts running the app instance.
func (a *App) Run() error {
	a.server.ServeSPA("/", a.Asset().embedded)
	a.server.ServeNoRoute()

	return a.command.Execute()
}
