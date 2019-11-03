package appy

import (
	"fmt"

	"github.com/bndr/gotabulate"
)

func newRoutesCommand(config *Config, logger *Logger, s *Server) *Cmd {
	return &Cmd{
		Use:   "routes",
		Short: "List all the server-side routes",
		Run: func(cmd *Cmd, args []string) {
			var routes [][]string
			for _, route := range s.Routes() {
				routes = append(routes, []string{route.Method, route.Path, route.Handler})
			}

			table := gotabulate.Create(routes)
			table.SetAlign("left")
			table.SetHeaders([]string{"Method", "Path", "Handler"})
			fmt.Println()
			fmt.Println(table.Render("simple"))
		},
	}
}