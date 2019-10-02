package cmd

import (
	"fmt"

	ah "github.com/appist/appy/http"

	"github.com/bndr/gotabulate"
	"github.com/spf13/cobra"
)

// NewRoutesCommand lists all the routes.
func NewRoutesCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "Lists all the routes.",
		Run: func(cmd *cobra.Command, args []string) {
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
