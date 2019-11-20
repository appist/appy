package appy

import (
	"html/template"
	"net/http"
	"os"

	"github.com/appist/appy/internal/cmd"
	appycmd "github.com/appist/appy/internal/cmd"
	appyhttp "github.com/appist/appy/internal/http"
	appymailer "github.com/appist/appy/internal/mailer"
	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

type (
	// App is the core of appy framework which determines how an application is driven.
	App struct {
		cmd       *Command
		config    *Config
		dbManager *DbManager
		logger    *Logger
		mailer    *Mailer
		server    *Server
	}

	// Command defines what a command line can do.
	Command = appycmd.Command

	// Config defines the application settings.
	Config = appysupport.Config

	// DbHandle is a database handle representing a pool of zero or more underlying connections. It's safe
	// for concurrent use by multiple goroutines.
	DbHandle = appyorm.DbHandle

	// DbHandleTx is an in-progress database transaction. It is safe for concurrent use by multiple goroutines.
	DbHandleTx = appyorm.DbHandleTx

	// DbManager manages multiple databases.
	DbManager = appyorm.DbManager

	// Logger provides the logging functionality.
	Logger = appysupport.Logger

	// Mailer provides the email sending via SMTP protocol.
	Mailer = appymailer.Mailer

	// Mail defines the email headers/body/attachments.
	Mail = appymailer.Mail

	// Context contains the HTTP request information.
	Context = appyhttp.Context

	// ContextKey is the HTTP context key with appy namespace.
	ContextKey = appyhttp.ContextKey

	// H is a type alias to map[string]string.
	H = appyhttp.H

	// Middleware is the middleware list attached to the server.
	Middleware = appyhttp.Middleware

	// HandlerFunc is a type alias to gin.HandlerFunc.
	HandlerFunc = appyhttp.HandlerFunc

	// RouterGroup can be used to group routes.
	RouterGroup = appyhttp.RouterGroup

	// RouteInfo provides the information about a route.
	RouteInfo = appyhttp.RouteInfo

	// Router manages the routing logic.
	Router = appyhttp.Router

	// Server is the engine that serves HTTP requests.
	Server = appyhttp.Server

	// AfterScanHook is the hook to trigger after a model's scan.
	AfterScanHook = appyorm.AfterScanHook

	// AfterSelectHook is the hook to trigger after a model's select.
	AfterSelectHook = appyorm.AfterSelectHook

	// BeforeInsertHook is the hook to trigger before a model's insert.
	BeforeInsertHook = appyorm.BeforeInsertHook

	// AfterInsertHook is the hook to trigger after a model's insert.
	AfterInsertHook = appyorm.AfterInsertHook

	// BeforeUpdateHook is the hook to trigger before a model's update.
	BeforeUpdateHook = appyorm.BeforeUpdateHook

	// AfterUpdateHook is the hook to trigger after a model's update.
	AfterUpdateHook = appyorm.AfterUpdateHook

	// BeforeDeleteHook is the hook to trigger before a model's delete.
	BeforeDeleteHook = appyorm.BeforeDeleteHook

	// AfterDeleteHook is the hook to trigger after a model's delete.
	AfterDeleteHook = appyorm.AfterDeleteHook
)

var (
	app *App

	// AESDecrypt decrypts a cipher text into a plain text using the key with AES.
	AESDecrypt = appysupport.AESDecrypt

	// AESEncrypt encrypts a plaintext into a cipher text using the key with AES.
	AESEncrypt = appysupport.AESEncrypt

	// ArrayContains checks if a value is in a slice of the same type.
	ArrayContains = appysupport.ArrayContains

	// CaptureOutput captures stdout and stderr.
	CaptureOutput = appysupport.CaptureOutput

	// DeepClone deeply clones from 1 interface to another.
	DeepClone = appysupport.DeepClone

	// IsCamelCase checks if a string is camelCase.
	IsCamelCase = appysupport.IsCamelCase

	// IsChainCase checks if a string is a chain-case.
	IsChainCase = appysupport.IsChainCase

	// IsFlatCase checks if a string is a flatcase.
	IsFlatCase = appysupport.IsFlatCase

	// IsPascalCase checks if a string is a PascalCase.
	IsPascalCase = appysupport.IsPascalCase

	// IsSnakeCase checks if a string is a snake_case.
	IsSnakeCase = appysupport.IsSnakeCase

	// IsDebugBuild indicates the current build is debug build which is meant for local development.
	IsDebugBuild = appysupport.IsDebugBuild

	// IsReleaseBuild indicates the current build is release build which is meant for production deployment.
	IsReleaseBuild = appysupport.IsReleaseBuild

	// IsConfigErrored is used to check if config contains any error during initialization.
	IsConfigErrored = appysupport.IsConfigErrored

	// IsProtectedEnv is used to protect the app from being destroyed by a command accidentally.
	IsProtectedEnv = appysupport.IsProtectedEnv

	// ParseEnv parses the environment variables into the config.
	ParseEnv = appysupport.ParseEnv

	// T translates a message based on the given key. Furthermore, we can pass in template data with `Count` in it to
	// support singular/plural cases.
	T = appyhttp.T

	// ToCamelCase converts a string to camelCase style.
	ToCamelCase = appysupport.ToCamelCase

	// ToChainCase converts a string to chain-case style.
	ToChainCase = appysupport.ToChainCase

	// ToFlatCase converts a string to flatcase style.
	ToFlatCase = appysupport.ToFlatCase

	// ToPascalCase converts a string to PascalCase style.
	ToPascalCase = appysupport.ToPascalCase

	// ToSnakeCase converts a string to snake_case style.
	ToSnakeCase = appysupport.ToSnakeCase
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}

// Init initializes App instance that comes with:
//
// - cmd: provides built-in commands and allow custom command constructing
//
// - config: provides configuration
//
// - logger: provides logger
//
// - mailer: provides the capability to send email via SMTP protocol.
//
// - server: provides the capability to serve HTTP requests
//
// - dbManager: manages the databases along with their pool connections
func Init(assets http.FileSystem, viewHelper template.FuncMap) {
	rootCmd := appycmd.NewCommand()
	logger := appysupport.NewLogger()
	config := appysupport.NewConfig(logger, assets)
	dbManager := appyorm.NewDbManager(logger)
	server := appyhttp.NewServer(config, logger, assets, viewHelper)
	server.InitSSR()
	mailer := appymailer.NewMailer(config, logger, server)

	if appysupport.IsDebugBuild() {
		rootCmd.AddCommand(
			cmd.NewBuildCommand(logger, server),
			cmd.NewStartCommand(logger, server),
		)
	}

	rootCmd.AddCommand(
		cmd.NewConfigDecryptCommand(config, logger),
		cmd.NewConfigEncryptCommand(config, logger),
		cmd.NewDbCreateCommand(config, dbManager, logger),
		cmd.NewDbDropCommand(config, dbManager, logger),
		cmd.NewGenMigrationCommand(config, dbManager, logger),
		cmd.NewDbMigrateCommand(config, dbManager, logger),
		cmd.NewDbMigrateStatusCommand(config, dbManager, logger),
		cmd.NewDbRollbackCommand(config, dbManager, logger),
		cmd.NewDbSchemaDumpCommand(config, dbManager, logger),
		cmd.NewDbSchemaLoadCommand(config, dbManager, logger),
		cmd.NewDbSeedCommand(config, dbManager, logger),
		cmd.NewDcDownCommand(logger, assets),
		cmd.NewDcRestartCommand(logger, assets),
		cmd.NewDcUpCommand(logger, assets),
		cmd.NewMiddlewareCommand(config, logger, server),
		cmd.NewRoutesCommand(config, logger, server),
		cmd.NewSecretCommand(logger),
		cmd.NewServeCommand(dbManager, logger, server),
		cmd.NewSetupCommand(config, dbManager, logger, assets),
		cmd.NewTeardownCommand(logger, assets),
		cmd.NewSSLSetupCommand(logger, server),
		cmd.NewSSLTeardownCommand(logger, server),
	)

	app = &App{
		cmd:       rootCmd,
		config:    config,
		dbManager: dbManager,
		logger:    logger,
		mailer:    mailer,
		server:    server,
	}
}

// Cmd returns the app's Cmd instance.
func (a App) Cmd() *appycmd.Command {
	return a.cmd
}

// Config returns the app's Config instance.
func (a App) Config() *appysupport.Config {
	return a.config
}

// DbManager returns the app's DbManager instance.
func (a App) DbManager() *appyorm.DbManager {
	return a.dbManager
}

// Logger returns the app's Logger instance.
func (a App) Logger() *appysupport.Logger {
	return a.logger
}

// Mailer returns the app's Mailer instance.
func (a App) Mailer() *appymailer.Mailer {
	return a.mailer
}

// Server returns the app's Server instance.
func (a App) Server() *appyhttp.Server {
	return a.server
}

// Run starts the application.
func (a App) Run() error {
	// Must be located right before the server runs due to CSR utilizes `NoRoute` to achieve pretty URL navigation
	// with HTML5 history API.
	a.server.InitCSR()

	// Start executing the root command.
	return a.Cmd().Execute()
}

// SetPlugins initializes the plugins.
func (a *App) SetPlugins(cb func(*App)) {
	cb(a)
}

// Default returns the app instance that is attached to appy module.
func Default() *App {
	return app
}
