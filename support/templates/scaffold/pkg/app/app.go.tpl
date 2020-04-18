package app

import (
	"{{.projectName}}/pkg/viewhelper"
	"path/filepath"
	"runtime"

	"github.com/appist/appy"
	"github.com/appist/appy/support"
)

type config struct {
	*appy.Config
	*appConfig
}

var (
	app *appy.App

	// Command is the application root command.
	Command *appy.Command

	// Config is the application config combined with appy's config.
	Config *config

	// I18n is the application I18n provider.
	I18n *appy.I18n

	// Mailer is the application mailer.
	Mailer *appy.Mailer

	// Logger is the application logger.
	Logger *appy.Logger

	// Server is the application server.
	Server *appy.Server

	// Worker is the application worker.
	Worker *appy.Worker
)

func init() {
	_, dirname, _, _ := runtime.Caller(0)
	appRoot, _ := filepath.Abs(dirname + "/../../..")
	app = appy.NewApp(asset, appRoot, viewhelper.New())

	// Setup the application's root command.
	Command = app.Command()
	Command.Short = "{{.projectDesc}}"

	// Setup the application's I18n provider.
	I18n = app.I18n()

	// Setup the application's logger.
	Logger = app.Logger()

	// Setup the application's mailer.
	Mailer = app.Mailer()

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

// DB returns the specific database handle.
func DB(name string) appy.DB {
	return app.DB(name)
}

// Run starts running the application.
func Run() {
	app.Run()
}
