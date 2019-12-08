package appy

import (
	"log"
	"net/http"
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
		assetsMngr *support.AssetsMngr
		command    *cmd.Command
		logger     *support.Logger
		server     *server.Server
	}
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// NewApp initializes an app instance.
func NewApp(static http.FileSystem) *App {
	command := cmd.NewCommand()
	logger := support.NewLogger()
	assetsMngr := support.NewAssetsMngr(nil, "", static)
	server := server.NewServer()

	return &App{
		assetsMngr: assetsMngr,
		command:    command,
		logger:     logger,
		server:     server,
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
