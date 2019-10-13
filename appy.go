package appy

import (
	"html/template"
	"net/http"

	"github.com/appist/appy/cmd"
	ah "github.com/appist/appy/http"
	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

// CommandT is a type alias to cobra.Command.
type CommandT = cobra.Command

// ConfigT offers a declarative way to map the environment variables.
type ConfigT = support.ConfigT

// ContextT is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type ContextT = ah.ContextT

// HandlerFuncT is a type alias to gin.HandlerFunc.
type HandlerFuncT = ah.HandlerFuncT

// RouterT is an alias to gin.Engine.
type RouterT = ah.RouterT

// RouterGroupT is an alias to gin.RouterGroup.
type RouterGroupT = ah.RouterGroupT

// RouteInfoT is an alias to gin.RouteInfo.
type RouteInfoT = ah.RouteInfoT

// RoutesT is an alias to gin.IRoutes.
type RoutesT = ah.RoutesT

// H is a type alias to map[string]interface{}.
type H map[string]interface{}

// LoggerT is a type alias to zap.SugaredLogger.
type LoggerT = support.LoggerT

// Config is the singleton that keeps the environment variables mapping defined in `support/config.go`.
var Config *support.ConfigT

// Logger is the singleton that provides logging utility to the app.
var Logger *support.LoggerT

// Server is the server singleton.
var Server *ah.ServerT

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
var DELETE func(relativePath string, handlers ...HandlerFuncT) RoutesT

// GET is a shortcut for router.Handle("GET", path, handle).
var GET func(relativePath string, handlers ...HandlerFuncT) RoutesT

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
var HEAD func(relativePath string, handlers ...HandlerFuncT) RoutesT

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
var OPTIONS func(relativePath string, handlers ...HandlerFuncT) RoutesT

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
var PATCH func(relativePath string, handlers ...HandlerFuncT) RoutesT

// POST is a shortcut for router.Handle("POST", path, handle).
var POST func(relativePath string, handlers ...HandlerFuncT) RoutesT

// PUT is a shortcut for router.Handle("PUT", path, handle).
var PUT func(relativePath string, handlers ...HandlerFuncT) RoutesT

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
var Any func(relativePath string, handlers ...HandlerFuncT) RoutesT

// BasePath returns the base path of router group.
// For example, if v := router.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".
var BasePath func() string

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
var Group func(relativePath string, handlers ...HandlerFuncT) *RouterGroupT

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
var Handle func(method, relativePath string, handlers ...HandlerFuncT) RoutesT

// HandleContext re-enter a context that has been rewritten.
// This can be done by setting c.Request.URL.Path to your new target.
// Disclaimer: You can loop yourself to death with this, use wisely.
var HandleContext func(ctx *ContextT)

// NoMethod adds handlers for NoMethod. It returns 405 code by default.
var NoMethod func(handlers ...HandlerFuncT)

// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
var NoRoute func(handlers ...HandlerFuncT)

// SecureJSONPrefix sets the secureJsonPrefix used in Context.SecureJSON.
var SecureJSONPrefix func(prefx string) *RouterT

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use :
//     router.Static("/static", "/var/www")
var Static func(relativePath, root string) RoutesT

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
var StaticFS func(relativePath string, fs http.FileSystem) RoutesT

// StaticFile registers a single route in order to serve a single file of the local filesystem.
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
var StaticFile func(relativePath, filePath string) RoutesT

// Use attaches a global middleware to the router. ie. the middleware attached though Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
var Use func(handlers ...HandlerFuncT) RoutesT

// T translates a message based on the given key. Furthermore, we can pass in template data with `Count` in it to
// support singular/plural cases.
var T = middleware.T

// ParseEnv parses the environment variables into the config.
var ParseEnv = support.ParseEnv

// Init initializes the server singleton.
func Init(assets http.FileSystem, viewHelper template.FuncMap) {
	support.Init(assets)
	Config = support.Config
	Logger = support.Logger

	Server = ah.NewServer(Config)
	Server.Assets = assets
	Server.InitSSR(viewHelper)

	DELETE = Server.Router.DELETE
	GET = Server.Router.GET
	HEAD = Server.Router.HEAD
	OPTIONS = Server.Router.OPTIONS
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
		cmd.AddCommand(cmd.NewBuildCommand(Server))
		cmd.AddCommand(cmd.NewStartCommand(Server))
		cmd.AddCommand(cmd.NewSSLCleanCommand(Server))
		cmd.AddCommand(cmd.NewSSLSetupCommand(Server))
	}
}

// Routes returns all the routes including those in middlewares.
func Routes() []RouteInfoT {
	return Server.Routes()
}

// Run executes the given command.
func Run() {
	// Shows a default welcome page with appy logo/slogan if `GET /` isn't defined.
	Server.AddDefaultWelcomePage()
	// Must be located right before the server runs due to CSR utilizes `gin.NoRoute` to achieve pretty URL navigation
	// with HTML5 history API.
	Server.InitCSR()

	cmd.Run()
}
