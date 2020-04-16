package cmd

import "github.com/appist/appy/support"

func newDcUpCommand(asset *support.Asset, logger *support.Logger) *Command {
	return &Command{
		Use:   "dc:up",
		Short: "Create and start containers that are defined in `.docker/docker-compose.yml`",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("up", asset)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}
}
