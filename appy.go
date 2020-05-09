package appy

import (
	"os"

	"github.com/appist/appy/cmd"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/appist/appy/worker"
)

var (
	// Build is the current build type for the application, can be "debug" or
	// "release". Please take note that this value will be updated to "release"
	// when running "go run . build" command.
	Build = support.Build

	// IsDebugBuild indicates the current build is debug build which is meant for
	// local development.
	IsDebugBuild = support.IsDebugBuild

	// IsReleaseBuild indicates the current build is release build which is meant
	// for production deployment.
	IsReleaseBuild = support.IsReleaseBuild

	// NewMockedWorkerHandler initializes a mocked WorkerHandler instance that is
	// useful for unit test.
	NewMockedWorkerHandler = worker.NewMockedHandler

	// NewJob initializes a job with a unique identifier and its data for
	// background job processing.
	NewJob = worker.NewJob

	// NewTestContext returns a fresh router w/ context for testing purposes.
	NewTestContext = pack.NewTestContext

	// NewTestLogger initializes a test Logger instance that is useful for testing
	// purpose.
	NewTestLogger = support.NewTestLogger

	// RunSuite takes a testing suite and runs all of the tests attached to it.
	RunSuite = test.Run

	// Scaffold creates a new application.
	Scaffold = support.Scaffold
)

type (
	// Asset manages the application assets.
	Asset = support.Asset

	// Command is used to build the command line interface.
	Command = cmd.Command

	// Config defines the application settings.
	Config = support.Config

	// Context contains the request information and is meant to be passed through
	// the entire HTTP request.
	Context = pack.Context

	// DB is the interface that manages the database config/connection/migrations.
	DB = record.DBer

	// DBManager manages the databases.
	DBManager = record.Engine

	// H is a shortcut for map[string]interface{}.
	H = support.H

	// HandlerFunc defines the handler used by middleware as return value.
	HandlerFunc = pack.HandlerFunc

	// I18n manages the application translations.
	I18n = support.I18n

	// Job represents a unit of work to be performed.
	Job = worker.Job

	// Logger provides the logging functionality.
	Logger = support.Logger

	// Mail defines the email headers/body/attachments.
	Mail = mailer.Mail

	// Mailer provides the capability to parse/render email template and send it
	// out via SMTP protocol.
	Mailer = mailer.Engine

	// Mock is the workhorse used to track activity on another object.
	Mock = test.Mock

	// ScaffoldOption contains the information of how a new application should be
	// created.
	ScaffoldOption = support.ScaffoldOption

	// Server processes the HTTP requests.
	Server = pack.Server

	// Suite is a basic testing suite with methods for storing and retrieving
	// the current *testing.T context.
	Suite = test.Suite

	// Tx is the interface that manages the database transaction.
	Tx = record.Txer

	// Worker processes the background jobs.
	Worker = worker.Engine

	// WorkerHandler processes background jobs.
	//
	// ProcessTask should return nil if the processing of a task is successful.
	//
	// If ProcessTask return a non-nil error or panics, the task will be retried
	// after delay.
	WorkerHandler = worker.Handler

	// WorkerHandlerFunc is an adapter to allow the use of ordinary functions as
	// a WorkerHandler. If f is a function with the appropriate signature,
	// WorkerHandlerFunc(f) is a WorkerHandler that calls f.
	WorkerHandlerFunc = worker.HandlerFunc
)

func init() {
	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}
}
