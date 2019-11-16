package cmd

import (
	"net/http"

	appysupport "github.com/appist/appy/internal/support"
)

// NewDcRestartCommand restart services that are defined in .docker/docker-compose.yml.
func NewDcRestartCommand(logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "dc:restart",
		Short: "Restart services that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("restart", assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}
