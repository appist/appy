//+build !test

package appy

func newDcDownCommand(asset *Asset, logger *Logger) *Command {
	return &Command{
		Use:   "dc:down",
		Short: "Stop and remove containers, networks, images, and volumes that are defined in `.docker/docker-compose.yml`",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("down", asset)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}
}
