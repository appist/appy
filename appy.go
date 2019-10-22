package appy

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/appist/appy/cmd"
	"github.com/appist/appy/core"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// App keeps everything that an application needs, e.g. config, logger, server and etc.
type App = core.App

// AppConfig keeps the parsed environment variables.
type AppConfig = core.AppConfig

// AppDb keeps database connection with its configuration.
type AppDb = core.AppDb

// AppDbConn represents a single database connection rather than a pool of database
// connections. Prefer running queries from DB unless there is a specific
// need for a continuous single database connection.
//
// A Conn must call Close to return the connection to the database pool
// and may do so concurrently with a running query.
//
// After a call to Close, all operations on the connection fail.
type AppDbConn = core.AppDbConn

// AppDbConfig keeps database connection options.
type AppDbConfig = core.AppDbConfig

// AppDbHandler is a database handle representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type AppDbHandler = core.AppDbHandler

// AppLogger keeps the logging functionality.
type AppLogger = core.AppLogger

// AppModel associates a type struct to a database table.
type AppModel = core.AppModel

// AppServer is the core that serves HTTP/GRPC requests.
type AppServer = core.AppServer

// Context retains the information that can be passed along in HTTP request flow.
type Context = core.Context

// H is a shortcut for map[string]interface{}.
type H = core.H

// HandlerFunc defines the handler used by middleware as the return value.
type HandlerFunc = core.HandlerFunc

// HandlersChain defines a HandlerFunc array.
type HandlersChain = core.HandlersChain

// Router keeps the rules that define how HTTP requests should be routed.
type Router = core.Router

// RouterGroup groups the routes together that share the same set of middlewares.
type RouterGroup = core.RouterGroup

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo = core.RouteInfo

// Routes defines all router handle interface.
type Routes = core.Routes

// Assert provides the unit test helpers to test various conditions.
type Assert = test.Assert

// TestSuite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
type TestSuite = test.Suite

var (
	app App

	// Build is the current build type for the application, can be `debug` or `release`.
	Build = core.Build

	// Config is the application's configuration singleton.
	Config AppConfig

	// Db is the application's database manager.
	Db map[string]*AppDb

	// Logger is the application's logger singleton.
	Logger *AppLogger

	// DELETE is a shortcut for appy.Handle("DELETE", path, handler).
	DELETE func(relativePath string, handlers ...HandlerFunc) Routes

	// GET is a shortcut for appy.Handle("GET", path, handler).
	GET func(relativePath string, handlers ...HandlerFunc) Routes

	// HEAD is a shortcut for appy.Handle("HEAD", path, handler).
	HEAD func(relativePath string, handlers ...HandlerFunc) Routes

	// OPTIONS is a shortcut for appy.Handle("OPTIONS", path, handler).
	OPTIONS func(relativePath string, handlers ...HandlerFunc) Routes

	// PATCH is a shortcut for appy.Handle("PATCH", path, handler).
	PATCH func(relativePath string, handlers ...HandlerFunc) Routes

	// POST is a shortcut for appy.Handle("POST", path, handler).
	POST func(relativePath string, handlers ...HandlerFunc) Routes

	// PUT is a shortcut for appy.Handle("PUT", path, handler).
	PUT func(relativePath string, handlers ...HandlerFunc) Routes

	// Any registers a route that matches all the HTTP methods.
	// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
	Any func(relativePath string, handlers ...HandlerFunc) Routes

	// BasePath returns the base path of router group.
	// For example, if v := appy.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".
	BasePath func() string

	// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
	// For example, all the routes that use a common middleware for authorization could be grouped.
	Group func(relativePath string, handlers ...HandlerFunc) *RouterGroup

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
	Handle func(method, relativePath string, handlers ...HandlerFunc) Routes

	// HandleContext re-enter a context that has been rewritten.
	// This can be done by setting c.Request.URL.Path to your new target.
	// Disclaimer: You can loop yourself to death with this, use wisely.
	HandleContext func(ctx *Context)

	// NoMethod adds handlers for NoMethod. It returns 405 code by default.
	NoMethod func(handlers ...HandlerFunc)

	// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
	NoRoute func(handlers ...HandlerFunc)

	// SecureJSONPrefix sets the secureJsonPrefix used in Context.SecureJSON.
	SecureJSONPrefix func(prefx string) *Router

	// Static serves files from the given file system root.
	// Internally a http.FileServer is used, therefore http.NotFound is used instead
	// of the Router's NotFound handler.
	// To use the operating system's file system implementation,
	// use: appy.Static("/static", "/var/www")
	Static func(relativePath, root string) Routes

	// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
	StaticFS func(relativePath string, fs http.FileSystem) Routes

	// StaticFile registers a single route in order to serve a single file of the local filesystem.
	// appy.StaticFile("favicon.ico", "./resources/favicon.ico")
	StaticFile func(relativePath, filePath string) Routes

	// Use attaches a global middleware to the router. ie. the middleware attached though Use() will be
	// included in the handlers chain for every single request. Even 404, 405, static files...
	// For example, this is the right place for a logger or error management middleware.
	Use func(handlers ...HandlerFunc) Routes

	// Middlewares is the global middlewares.
	Middlewares HandlersChain

	// NewAssert returns an Assert instance that provides the unit test helpers to test various conditions.
	NewAssert = test.NewAssert

	// ArrayContains checks if a value is in a slice of the same type.
	ArrayContains = support.ArrayContains

	// CaptureOutput captures stdout and stderr.
	CaptureOutput = core.CaptureOutput

	// DeepClone deeply clones from 1 interface to another.
	DeepClone = support.DeepClone

	// ParseEnv parses the environment variables into the config.
	ParseEnv = support.ParseEnv

	// T translates a message based on the given key. Furthermore, we can pass in template data with `Count` in it to
	// support singular/plural cases.
	T = core.T
)

// CaptureLoggerOutput captures the Logger's output.
func CaptureLoggerOutput(f func()) string {
	var buffer bytes.Buffer
	oldLogger := Logger
	writer := bufio.NewWriter(&buffer)
	Logger = &AppLogger{
		SugaredLogger: zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(writer),
				zapcore.DebugLevel,
			),
		).Sugar(),
	}
	f()
	writer.Flush()
	Logger = oldLogger

	return buffer.String()
}

// Init initializes the application singleton.
func Init(assets http.FileSystem, appConf interface{}, viewHelper template.FuncMap) {
	var err error
	app, err = core.NewApp(assets, appConf, viewHelper)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Initialize the SSR rendering pipeline.
	app.Server.InitSSR()

	Config = app.Config
	Db = app.Db
	Logger = app.Logger
	DELETE = app.Server.Router.DELETE
	GET = app.Server.Router.GET
	HEAD = app.Server.Router.HEAD
	OPTIONS = app.Server.Router.OPTIONS
	PATCH = app.Server.Router.PATCH
	POST = app.Server.Router.POST
	PUT = app.Server.Router.PUT
	Any = app.Server.Router.Any
	BasePath = app.Server.Router.BasePath
	Group = app.Server.Router.Group
	Handle = app.Server.Router.Handle
	HandleContext = app.Server.Router.HandleContext
	NoMethod = app.Server.Router.NoMethod
	NoRoute = app.Server.Router.NoRoute
	SecureJSONPrefix = app.Server.Router.SecureJsonPrefix
	Static = app.Server.Router.Static
	StaticFS = app.Server.Router.StaticFS
	StaticFile = app.Server.Router.StaticFile
	Use = app.Server.Router.Use
	Middlewares = app.Server.Router.Handlers

	cmd.Init(app)
	cmd.AddCommand(cmd.NewDbCreateCommand(app.Config, app.Db))
	cmd.AddCommand(cmd.NewDbDropCommand(app.Config, app.Db))
	cmd.AddCommand(cmd.NewMiddlewareCommand(app.Server))
	cmd.AddCommand(cmd.NewRoutesCommand(app.Server))
	cmd.AddCommand(cmd.NewSecretCommand())
	cmd.AddCommand(cmd.NewServeCommand(app.Server))

	if Build != "release" {
		cmd.AddCommand(cmd.NewBuildCommand(app.Server))
		cmd.AddCommand(cmd.NewConfigDecryptCommand(app.Server))
		cmd.AddCommand(cmd.NewConfigEncryptCommand(app.Server))
		cmd.AddCommand(cmd.NewStartCommand(app.Server))
		cmd.AddCommand(cmd.NewSSLCleanCommand(app.Server))
		cmd.AddCommand(cmd.NewSSLSetupCommand(app.Server))
	}
}

// Run executes the given command.
func Run() {
	// Shows a default welcome page with appy logo/slogan if `GET /` isn't defined.
	app.Server.AddDefaultWelcomePage()

	// Must be located right before the server runs due to CSR utilizes `NoRoute` to achieve pretty URL navigation
	// with HTML5 history API. In addition, we only enable CSR hosting for `release` build due to `debug` build
	// should rely on webpack-dev-server.
	if Build == "release" {
		app.Server.InitCSR()
	}

	for _, db := range app.Db {
		err := db.Connect()
		if err != nil {
			Logger.Fatal(err)
		}

		defer db.Close()
	}

	cmd.Run()
}
