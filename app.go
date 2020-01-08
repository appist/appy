package appy

import (
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
		logger *Logger
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp() *App {
	logger := NewLogger()

	return &App{
		logger: logger,
	}
}

// Logger returns the app instance's logger.
func (a *App) Logger() *Logger {
	return a.logger
}
