package cmd

import (
	"fmt"

	appyhttp "github.com/appist/appy/internal/http"
	appysupport "github.com/appist/appy/internal/support"
	"github.com/bndr/gotabulate"
)

// NewRoutesCommand list all the server-side routes.
func NewRoutesCommand(config *appysupport.Config, logger *appysupport.Logger, s *appyhttp.Server) *Command {
	return &Command{
		Use:   "routes",
		Short: "List all the server-side routes",
		Run: func(cmd *Command, args []string) {
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
