package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
)

func newMiddlewareCommand(config *support.Config, logger *support.Logger, server *pack.Server) *Command {
	return &Command{
		Use:   "middleware",
		Short: "List all the global middleware",
		Run: func(cmd *Command, args []string) {
			regex := regexp.MustCompile(`\.func.*`)

			for _, mw := range server.Middleware() {
				p := reflect.ValueOf(mw).Pointer()
				f := runtime.FuncForPC(p)
				fmt.Println(regex.ReplaceAllString(f.Name(), ""))
			}
		},
	}
}
