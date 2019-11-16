package cmd

import (
	"net/http"

	appysupport "github.com/appist/appy/internal/support"
)

// NewDcUpCommand create and start containers that are defined in .docker/docker-compose.yml.
func NewDcUpCommand(logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "dc:up",
		Short: "Create and start containers that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("up", assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}
