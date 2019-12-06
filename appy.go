package appy

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/appist/appy/cmd"
	"github.com/appist/appy/server"
	"github.com/appist/appy/support"
)

type (
	// App is the framework core that drives the application.
	App struct {
		command *cmd.Command
		logger  *support.Logger
		server  *server.Server
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp() *App {
	command := cmd.NewCommand()
	logger := support.NewLogger()
	server := server.NewServer()

	return &App{
		command: command,
		logger:  logger,
		server:  server,
	}
}

// Run starts running the app instance.
func (a App) Run() error {
	return a.command.Execute()
}

// Bootstrap initializes the project layout.
func Bootstrap() {
	_, path, _, _ := runtime.Caller(0)
	appTplPath := filepath.Dir(path) + "/templates/app"

	err := filepath.Walk(appTplPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
