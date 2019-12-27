package cmd

import (
	"github.com/appist/appy/support"
)

// NewDcRestartCommand restart services that are defined in `.docker/docker-compose.yml`.
func NewDcRestartCommand(assets *support.Assets, logger *support.Logger) *Command {
	cmd := &Command{
		Use:   "dc:restart",
		Short: "Restart services that are defined in `.docker/docker-compose.yml`",
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
