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

	return &App{
		asset:   asset,
		config:  config,
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

// Logger returns the app instance's logger.
func (a *App) Logger() *Logger {
	return a.logger
}
