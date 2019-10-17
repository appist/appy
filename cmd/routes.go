package cmd

import (
	"fmt"

	"github.com/appist/appy/core"
	"github.com/bndr/gotabulate"
)

// NewRoutesCommand lists all the routes.
func NewRoutesCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "routes",
		Short: "Lists all the routes.",
		Run: func(cmd *AppCmd, args []string) {
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
