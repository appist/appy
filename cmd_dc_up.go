//+build !test

package appy

func newDcUpCommand(asset *Asset, logger *Logger) *Command {
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
