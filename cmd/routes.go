package cmd

import (
	"fmt"

	ah "github.com/appist/appy/http"

	"github.com/bndr/gotabulate"
	"github.com/spf13/cobra"
)

// NewHTTPRoutesCommand lists all the HTTP routes.
func NewHTTPRoutesCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "http:routes",
		Short: "Lists all the HTTP routes.",
		Run: func(cmd *cobra.Command, args []string) {
			var routes [][]string

			for _, route := range s.GetAllRoutes() {
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
