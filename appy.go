package appy

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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
	// value will be updated to `release` when running `go run . build` command.
	Build = DebugBuild
)

type (
	// H is a shortcut for map[string]interface{}.
	H map[string]interface{}
)

// IsDebugBuild indicates the current build is debug build which is meant for local development.
func IsDebugBuild() bool {
	return Build == DebugBuild
}

// IsReleaseBuild indicates the current build is release build which is meant for production deployment.
func IsReleaseBuild() bool {
	return Build == ReleaseBuild
}

// Scaffold generates a new project using the template.
func Scaffold(name, description string) {
	_, dirname, _, _ := runtime.Caller(0)
	tplPath := filepath.Dir(dirname) + "/templates/scaffold"

	err := filepath.Walk(tplPath,
		func(src string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			dest := strings.ReplaceAll(src, tplPath+"/", "")
			if info.IsDir() {
				err := os.MkdirAll(dest, 0777)
				if err != nil {
					log.Fatal(err)
				}
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
