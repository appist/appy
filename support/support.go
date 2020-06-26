package support

import (
	"database/sql/driver"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	// DebugBuild tends to be slow as it includes debug lvl logging which is
	// more verbose.
	DebugBuild = "debug"

	// ReleaseBuild tends to be faster as it excludes debug lvl logging.
	ReleaseBuild = "release"

	// VERSION follows semantic versioning to indicate the framework's release
	// status.
	VERSION = "0.1.0"

	// DESCRIPTION indicates what the web framework is aiming to provide.
	DESCRIPTION = "An opinionated productive web framework that helps scaling business easier."
)

var (
	// Build is the current build type for the application, can be "debug" or
	// "release". Please take note that this value will be updated to "release"
	// when running "go run . build" command.
	Build = DebugBuild

	// SupportedDBAdapters indicates the list of database adapters that are
	// supported.
	SupportedDBAdapters = []string{"mysql", "postgres"}
)

type (
	// H is a shortcut for map[string]interface{}.
	H map[string]interface{}
)

func init() {
	recordTypes := []interface{}{
		NBool{},
		NFloat64{},
		NInt64{},
		NString{},
		NTime{},
		ZBool{},
		ZFloat64{},
		ZInt64{},
		ZString{},
		ZTime{},
	}

	ginValidator, _ := binding.Validator.Engine().(*validator.Validate)
	ginValidator.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		var val driver.Value

		if valuer, ok := field.Interface().(driver.Valuer); ok {
			v, err := valuer.Value()

			if err == nil {
				val = v
			}
		}

		return val
	}, recordTypes...)
}

// IsDebugBuild indicates the current build is debug build which is meant for
// local development.
func IsDebugBuild() bool {
	return Build == DebugBuild
}

// IsReleaseBuild indicates the current build is release build which is meant
// for production deployment.
func IsReleaseBuild() bool {
	return Build == ReleaseBuild
}
