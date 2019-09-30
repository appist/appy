package appy

import (
	"net/http"

	"github.com/appist/appy/cmd"
	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ContextT is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type ContextT = gin.Context

// CommandT is a type alias to cobra.Command.
type CommandT = cobra.Command

// HandlerFuncT is a type alias to gin.HandlerFunc.
type HandlerFuncT = gin.HandlerFunc

type RouterT = ah.RouterT
type RouterGroupT = ah.RouterGroupT
type RoutesT = ah.RoutesT

// H is a type alias to map[string]interface{}.
type H map[string]interface{}

// LoggerT is a type alias to zap.SugaredLogger.
type LoggerT = support.LoggerT

// Config is the singleton that keeps the environment variables mapping defined in `support/config.go`.
var Config = support.Config

// Logger is the singleton that provides logging utility to the app.
var Logger = support.Logger

// Server is the server singleton.
var Server *ah.ServerT

var DELETE func(relativePath string, handlers ...HandlerFuncT) RoutesT
var GET func(relativePath string, handlers ...HandlerFuncT) RoutesT
var HEAD func(relativePath string, handlers ...HandlerFuncT) RoutesT
var PATCH func(relativePath string, handlers ...HandlerFuncT) RoutesT
var POST func(relativePath string, handlers ...HandlerFuncT) RoutesT
var PUT func(relativePath string, handlers ...HandlerFuncT) RoutesT
var Any func(relativePath string, handlers ...HandlerFuncT) RoutesT
var BasePath func() string
var Group func(relativePath string, handlers ...HandlerFuncT) *RouterGroupT
var Handle func(method, relativePath string, handlers ...HandlerFuncT) RoutesT
var HandleContext func(ctx *ContextT)
var NoMethod func(handlers ...HandlerFuncT)
var NoRoute func(handlers ...HandlerFuncT)
var SecureJSONPrefix func(prefx string) *RouterT
var Static func(relativePath, root string) RoutesT
var StaticFS func(relativePath string, fs http.FileSystem) RoutesT
var StaticFile func(relativePath, filePath string) RoutesT
var Use func(handlers ...HandlerFuncT) RoutesT

// Init initializes the server singleton.
func Init(assets http.FileSystem) {
	Server = ah.NewServer(Config)
	DELETE = Server.Router.DELETE
	GET = Server.Router.GET
	HEAD = Server.Router.HEAD
	PATCH = Server.Router.PATCH
	POST = Server.Router.POST
	PUT = Server.Router.PUT
	Any = Server.Router.Any
	BasePath = Server.Router.BasePath
	Group = Server.Router.Group
	Handle = Server.Router.Handle
	HandleContext = Server.Router.HandleContext
	NoMethod = Server.Router.NoMethod
	NoRoute = Server.Router.NoRoute
	SecureJSONPrefix = Server.Router.SecureJsonPrefix
	Static = Server.Router.Static
	StaticFS = Server.Router.StaticFS
	StaticFile = Server.Router.StaticFile
	Use = Server.Router.Use

	cmd.AddCommand(cmd.NewRoutesCommand(Server))
	cmd.AddCommand(cmd.NewServeCommand(Server))
	cmd.AddCommand(cmd.NewSecretCommand())

	if support.Build != "release" {
		// cmd.AddCommand(cmd.NewBuildCommand(Server))
		// cmd.AddCommand(cmd.NewHTTPDevCommand(Server))
		cmd.AddCommand(cmd.NewSSLCleanCommand(Server))
		cmd.AddCommand(cmd.NewSSLSetupCommand(Server))
	}
}

func Routes() []ah.RouteInfoT {
	return Server.Routes()
}

// Run executes the given command.
func Run() {
	cmd.Run()
}
