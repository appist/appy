package appy

import (
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
		asset      *Asset
		config     *Config
		i18n       *I18n
		logger     *Logger
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

	// Setup the default middleware.
	server.Use(AttachLogger(logger))
	server.Use(AttachI18n(i18n))
	server.Use(AttachViewEngine(asset, config, logger, viewFuncs))
	server.Use(RealIP())
	server.Use(RequestID())
	server.Use(RequestLogger(config, logger))
	server.Use(Gzip(config))
	server.Use(HealthCheck(config.HTTPHealthCheckURL))
	server.Use(CSRF(config, logger, support))
	server.Use(Secure(config))
	server.Use(APIOnlyResponse())
	server.Use(SessionManager(config))
	server.Use(Recovery(logger))

	return &App{
		asset:      asset,
		config:     config,
		i18n:       i18n,
		logger:     logger,
		server:     server,
		support:    support,
		viewEngine: viewEngine,
	}
}

// Asset returns the app instance's asset.
func (a *App) Asset() *Asset {
	return a.asset
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

// Server returns the app instance's server.
func (a *App) Server() *Server {
	return a.server
}

// ViewEngine returns the app instance's view engine.
func (a *App) ViewEngine() *ViewEngine {
	return a.viewEngine
}
