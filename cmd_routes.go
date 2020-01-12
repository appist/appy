//+build !test

package appy

import (
	"fmt"

	"github.com/bndr/gotabulate"
)

func newRoutesCommand(config *Config, logger *Logger, server *Server) *Command {
	return &Command{
		Use:   "routes",
		Short: "List all the server-side routes",
		Run: func(cmd *Command, args []string) {
			var routes [][]string
			for _, route := range server.Routes() {
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
