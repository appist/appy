package core

import (
	"html/template"
	"net/http"
)

const (
	// VERSION follows semantic versioning to indicate the framework's release status.
	VERSION = "0.1.0"
)

// App keeps everything that an application needs, e.g. config, logger, server and etc.
type App struct {
	Config AppConfig
	Logger *AppLogger
	Server Server
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

	server := newServer(assets, config, logger, viewHelper)
	if err != nil {
		return app, err
	}

	app.Config = config
	app.Logger = logger
	app.Server = server
	return app, nil
}
