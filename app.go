package appy

type (
	App struct {
		cmd       *Cmd
		config    *Config
		logger    *Logger
		server    *Server
		dbManager *DbManager
		support   *Support
	}
)

var (
	// Build is the current build type for the application, can be `debug` or `release`.
	Build = "debug"
)

// NewApp initializes App instance that comes with:
//
// cmd - provides appy's built-in commands and allow custom command constructing
// config - provides appy's global configuration
// logger - provides logger
// server - provides the capability to serve HTTP/GRPC requests
// dbManager - manages the databases along with their pool connections
// support - provides utility helpers/extensions
func NewApp() App {
	app := App{
		logger:  NewLogger(Build),
		support: NewSupport(),
	}

	return app
}

// Cmd returns the app's Cmd instance.
func (a App) Cmd() *Cmd {
	return a.cmd
}

// Config returns the app's Config instance.
func (a App) Config() *Config {
	return a.config
}

// DbManager returns the app's DbManager instance.
func (a App) DbManager() *DbManager {
	return a.dbManager
}

// Logger returns the app's Logger instance.
func (a App) Logger() *Logger {
	return a.logger
}

// Server returns the app's Server instance.
func (a App) Server() *Server {
	return a.server
}

// Support returns the app's Support instance.
func (a App) Support() *Support {
	return a.support
}

func (a App) Run() {

}
