package support

import "errors"

const (
	// DebugBuild tends to be slow as it includes debug lvl logging which is more verbose.
	DebugBuild = "debug"

	// ReleaseBuild tends to be faster as it excludes debug lvl logging.
	ReleaseBuild = "release"

	// VERSION follows semantic versioning to indicate the framework's release status.
	VERSION = "0.1.0"

	// DESCRIPTION indicates what the web framework is aiming to provide.
	DESCRIPTION = "An opinionated productive web framework that helps scaling business easier."
)

var (
	// Build is the current build type for the application, can be `debug` or `release`. Please take note that this
	// value will be updated to `release` by `go run . build` command.
	Build = DebugBuild

	// ErrNoMasterKey indicates the master key is not provided.
	ErrNoMasterKey = errors.New("master key is not provided")

	// ErrReadMasterKeyFile indicates there is a problem reading master key file.
	ErrReadMasterKeyFile = errors.New("failed to read master key file")

	// ErrNoConfigInAssets indicates the config is missing in the assets.
	ErrNoConfigInAssets = errors.New("missing config in the assets")
)

// IsDebugBuild indicates the current build is debug build which is meant for local development.
func IsDebugBuild() bool {
	return Build == DebugBuild
}

// IsReleaseBuild indicates the current build is release build which is meant for production deployment.
func IsReleaseBuild() bool {
	return Build == DebugBuild
}
