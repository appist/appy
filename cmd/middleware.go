package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
)

// NewMiddlewareCommand list all the global middleware.
func NewMiddlewareCommand(config *support.Config, logger *support.Logger, server *ah.Server) *Command {
	cmd := &Command{
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

	return cmd
}
