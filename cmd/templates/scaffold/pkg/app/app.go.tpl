package app

import (
	"{{.projectName}}/pkg/viewhelper"
	"path/filepath"
	"runtime"

	"github.com/appist/appy"
	"github.com/appist/appy/cmd"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/worker"
)

type config struct {
	*support.Config
	*appConfig
}

var (
	app *appy.App

	// Command is the application root command.
	Command *cmd.Command

	// Config is the application config combined with appy's config.
	Config *config

	// DB returns the database of the specified name.
	DB func(name string) record.DBer

	// DBManager is the application's DB manager.
	DBManager *record.Engine

	// I18n is the application I18n provider.
	I18n *support.I18n

	// Logger is the application logger.
	Logger *support.Logger

	// Mailer is the application mailer.
	Mailer *mailer.Engine

	// Model returns the layer that represents business data and logic.
	Model func(dest interface{}, opts ...record.ModelOption) record.Modeler

	// Server is the application server.
	Server *pack.Server

	// Worker is the application worker.
	Worker *worker.Engine
)

func init() {
	_, dirname, _, _ := runtime.Caller(0)
	appRoot, _ := filepath.Abs(dirname + "/../../..")
	app = appy.NewApp(asset, appRoot, viewhelper.New())

	// Setup the application's root command.
	Command = app.Command()
	Command.Short = "{{.projectDesc}}"

	// Setup the DB function alias.
	DB = app.DB

	// Setup the application's DB manager.
	DBManager = app.DBManager()

	// Setup the application's I18n provider.
	I18n = app.I18n()

	// Setup the application's logger.
	Logger = app.Logger()

	// Setup the application's mailer.
	Mailer = app.Mailer()

	// Setup the Model function alias.
	Model = app.Model

	// Setup the application's server.
	Server = app.Server()

	// Setup the application's worker.
	Worker = app.Worker()

	// Setup the application's config.
	c := &appConfig{}
	err := support.ParseEnv(c)
	if err != nil {
		Logger.Fatal(err)
	}

	Config = &config{
		app.Config(),
		c,
	}
}

// Run starts running the application.
func Run() {
	app.Run()
}
