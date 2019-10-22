package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"

	"github.com/appist/appy/core"
)

// NewMiddlewareCommand prints out the middleware list.
func NewMiddlewareCommand(s core.AppServer) *AppCmd {
	cmd := &AppCmd{
		Use:   "middleware",
		Short: "Prints out the middleware list.",
		Run: func(cmd *AppCmd, args []string) {
			regex := regexp.MustCompile(`\.func.*`)

			for _, mw := range s.Router.Handlers {
				p := reflect.ValueOf(mw).Pointer()
				f := runtime.FuncForPC(p)
				fmt.Println(regex.ReplaceAllString(f.Name(), ""))
			}
		},
	}

	return cmd
}
