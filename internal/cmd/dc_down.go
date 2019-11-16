package cmd

import (
	"net/http"

	appysupport "github.com/appist/appy/internal/support"
)

// NewDcDownCommand stop and remove containers, networks, images, and volumes that are defined in .docker/docker-compose.yml.
func NewDcDownCommand(logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "dc:down",
		Short: "Stop and remove containers, networks, images, and volumes that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("down", assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}
