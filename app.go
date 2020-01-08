package appy

import (
	"net/http"
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
		asset  *Asset
		logger *Logger
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp(embedded http.FileSystem) *App {
	asset := NewAsset(embedded, nil)
	logger := NewLogger()

	return &App{
		asset:  asset,
		logger: logger,
	}
}

// Asset returns the app instance's asset.
func (a *App) Asset() *Asset {
	return a.asset
}

// Logger returns the app instance's logger.
func (a *App) Logger() *Logger {
	return a.logger
}
