package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"

	"github.com/appist/appy/core"
)

// NewMiddlewareCommand lists all the middlewares.
func NewMiddlewareCommand(s core.AppServer) *AppCmd {
	cmd := &AppCmd{
		Use:   "middleware",
		Short: "Lists all the middlewares.",
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
