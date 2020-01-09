package appy

import (
	"net/http"
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
		asset   *Asset
		config  *Config
		i18n    *I18n
		logger  *Logger
		support Supporter
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp(embedded http.FileSystem) *App {
	support := &Support{}
	asset := NewAsset(embedded, nil)
	logger := NewLogger()
	config := NewConfig(asset, logger, support)
	i18n := NewI18n(asset, config, logger)

	return &App{
		asset:   asset,
		config:  config,
		i18n:    i18n,
		logger:  logger,
		support: support,
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
