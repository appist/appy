package appy

import (
	"html/template"
	"net/http"

	"appist/appy/cmd"
	"appist/appy/database"
	ahttp "appist/appy/http"
	"appist/appy/middleware"
	"appist/appy/support"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
	"github.com/spf13/cobra"
)

// CommandT is a type alias to cobra.Command.
type CommandT = cobra.Command

// ContextT is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type ContextT = gin.Context

// H is a type alias to map[string]interface{}.
type H map[string]interface{}

// HandlerFuncT is a type alias to gin.HandlerFunc.
type HandlerFuncT = gin.HandlerFunc

// LoggerT is a type alias to zap.SugaredLogger.
type LoggerT = support.LoggerT

// SessionT is a type alias to appy's session type that wraps up gorilla/gin session.
type SessionT = middleware.Session

// HTTPServer is the singleton that serves HTTP requests.
var HTTPServer *ahttp.ServerT

// Config is the singleton that keeps the environment variables mapping defined in `support/config.go`.
var Config = support.Config

// Db is the database connection handlers.
var Db = map[string]*pg.DB{}

// DbConfigs is the config for database connection handlers.
var DbConfigs map[string]*pg.Options

// Logger is the singleton that provides logging utility to the app.
var Logger = support.Logger

// Bootstrap initializes the app instance with singletons like Config, Logger and etc.
func Bootstrap(assets http.FileSystem, fm template.FuncMap) {
	HTTPServer = ahttp.NewServer(Config)
	HTTPServer.SetupAssets(assets)
	HTTPServer.SetupI18n()
	HTTPServer.SetFuncMap(fm)

	DbConfigs = database.ParseDbConfigs()

	cmd.AddCommand(cmd.NewHTTPRoutesCommand(HTTPServer))
	cmd.AddCommand(cmd.NewHTTPServeCommand(HTTPServer))
	cmd.AddCommand(cmd.NewSecretCommand())

	if support.Build != "release" {
		cmd.AddCommand(cmd.NewBuildCommand(HTTPServer))
		cmd.AddCommand(cmd.NewHTTPDevCommand(HTTPServer))
		cmd.AddCommand(cmd.NewSSLCleanCommand())
		cmd.AddCommand(cmd.NewSSLSetupCommand())
	}
}

// Run executes the given command.
func Run() {
	for name, conf := range DbConfigs {
		Db[name] = pg.Connect(conf)
		defer Db[name].Close()
	}

	cmd.Run()
}
