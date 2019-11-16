package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"

	appyhttp "github.com/appist/appy/internal/http"
	appysupport "github.com/appist/appy/internal/support"
)

// NewMiddlewareCommand list all the middleware.
func NewMiddlewareCommand(config *appysupport.Config, logger *appysupport.Logger, s *appyhttp.Server) *Command {
	cmd := &Command{
		Use:   "middleware",
		Short: "List all the middleware",
		Run: func(cmd *Command, args []string) {
			regex := regexp.MustCompile(`\.func.*`)

			for _, mw := range s.Middleware() {
				p := reflect.ValueOf(mw).Pointer()
				f := runtime.FuncForPC(p)
				fmt.Println(regex.ReplaceAllString(f.Name(), ""))
			}
		},
	}

	return cmd
}
