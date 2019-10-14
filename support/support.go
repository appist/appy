package support

import (
	"net/http"
	"regexp"
)

// VERSION is the current version of appy.
const VERSION = "0.1.0"

var (
	// Config is the singleton that keeps the environment variables mapping defined in `support/config.go`.
	Config *ConfigT

	// Logger is the singleton that provides logging utility to the app.
	Logger *LoggerT

	// ExtRegex is the regular expression to match assets path.
	ExtRegex = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
)

// Init setup the Config and Logger singletons.
func Init(assets http.FileSystem, ssrRoot string) {
	Config, _ = NewConfig(assets, ssrRoot)
	Logger, _ = NewLogger(NewLoggerConfig())
}
