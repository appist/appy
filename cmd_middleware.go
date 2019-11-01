package appy

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

func newMiddlewareCommand(config *Config, logger *Logger, s *Server) *Cmd {
	cmd := &Cmd{
		Use:   "middleware",
		Short: "Lists all the middlewares",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			regex := regexp.MustCompile(`\.func.*`)

			for _, mw := range s.router.Handlers {
				p := reflect.ValueOf(mw).Pointer()
				f := runtime.FuncForPC(p)
				fmt.Println(regex.ReplaceAllString(f.Name(), ""))
			}
		},
	}

	return cmd
}
