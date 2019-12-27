package cmd

import (
	"github.com/appist/appy/support"
)

// NewDcUpCommand create and start containers that are defined in `.docker/docker-compose.yml`.
func NewDcUpCommand(assets *support.Assets, logger *support.Logger) *Command {
	cmd := &Command{
		Use:   "dc:up",
		Short: "Create and start containers that are defined in `.docker/docker-compose.yml`",
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
