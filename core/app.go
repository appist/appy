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
	isConfigRequired := false
	// Keeps track of the commands that require master key & config check.
	configRequiredCmds := []string{"db:create", "db:drop", "db:migrate", "db:migrate:status", "db:rollback", "db:seed", "serve", "start"}

	for _, configRequiredCmd := range configRequiredCmds {
		if support.ArrayContains(os.Args, configRequiredCmd) {
			isConfigRequired = true
			break
		}
	}

	logger, err := newLogger(newLoggerConfig())
	if err != nil {
		return app, err
	}
	app.Logger = logger

	masterKey, err := MasterKey()
	if err != nil && isConfigRequired && len(os.Args) > 1 {
		return app, err
	}

	config, dbConfig, err := newConfig(assets, appConf, masterKey, logger)
	if err != nil {
		if isConfigRequired && len(os.Args) > 1 {
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
		app.Db[name], err = newDb(val, app.Logger)
		if err != nil {
			return app, err
		}
	}

	return app, nil
}
