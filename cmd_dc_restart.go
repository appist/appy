//+build !test

package appy

func newDcRestartCommand(asset *Asset, logger *Logger) *Command {
	return &Command{
		Use:   "dc:restart",
		Short: "Restart services that are defined in `.docker/docker-compose.yml`",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("restart", asset)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}
}
