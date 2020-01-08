package appy

import (
	"os"
)

type (
	// App is the framework core that drives the application.
	App struct {
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp() *App {
	return &App{}
}
