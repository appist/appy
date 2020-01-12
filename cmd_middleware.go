//+build !test

package appy

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

func newMiddlewareCommand(config *Config, logger *Logger, server *Server) *Command {
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
