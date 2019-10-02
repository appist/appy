package support

// VERSION is the current version of appy.
const VERSION = "0.1.0"

var (
	// Config is the singleton that keeps the environment variables mapping defined in `support/config.go`.
	Config *ConfigT
	// Logger is the singleton that provides logging utility to the app.
	Logger *LoggerT
)

func init() {
	Config, _ = NewConfig()
	Logger, _ = NewLogger(NewLoggerConfig())
}
