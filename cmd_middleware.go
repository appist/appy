package appy

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

func newMiddlewareCommand(s *Server) *Cmd {
	cmd := &Cmd{
		Use:   "middleware",
		Short: "Lists all the middlewares",
		Run: func(cmd *Cmd, args []string) {
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
