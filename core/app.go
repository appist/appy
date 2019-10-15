package core

import (
	"html/template"
	"net/http"
)

type App struct {
	Config AppConfig
	Logger *SugaredLogger
	server server
}

// NewApp initializes the app singleton.
func NewApp(assets http.FileSystem, viewHelper template.FuncMap) (App, error) {
	app := App{}
	config, err := newConfig(assets)
	if err != nil {
		return app, err
	}

	logger, err := newLogger(newLoggerConfig())
	if err != nil {
		return app, err
	}

	server := newServer(assets, viewHelper)
	if err != nil {
		return app, err
	}

	app.Config = config
	app.Logger = logger
	app.server = server
	return app, nil
}
