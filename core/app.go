package core

import (
	"html/template"
	"net/http"
	"os"

	"github.com/appist/appy/support"
)

const (
	// VERSION follows semantic versioning to indicate the framework's release status.
	VERSION = "0.1.0"
)

// App keeps everything that an application needs, e.g. config, logger, server and etc.
type App struct {
	Config AppConfig
	Db     map[string]*AppDb
	Logger *AppLogger
	Server AppServer
}

// NewApp initializes the app singleton.
func NewApp(assets http.FileSystem, appConf interface{}, viewHelper template.FuncMap) (App, error) {
	app := App{}
	logger, err := newLogger(newLoggerConfig())
	if err != nil {
		return app, err
	}
	app.Logger = logger

	config, dbConfig, err := newConfig(assets, appConf, logger)
	if err != nil {
		isConfigOptional := false
		configOptionalCmds := []string{"build", "config:dec", "config:enc", "middleware", "routes", "secret", "ssl:clean", "ssl:setup", "start", "-h", "--help"}

		for _, configOptionalCmd := range configOptionalCmds {
			if support.ArrayContains(os.Args, configOptionalCmd) {
				isConfigOptional = true
				break
			}
		}

		if !isConfigOptional && len(os.Args) > 1 {
			return app, err
		}

		err = nil
	}
	app.Config = config

	server := newServer(assets, app.Config, app.Logger, viewHelper)
	if err != nil {
		return app, err
	}
	app.Server = server

	app.Db = map[string]*AppDb{}
	for name, val := range dbConfig {
		app.Db[name], err = newDb(val)
		if err != nil {
			return app, err
		}
	}

	return app, nil
}
