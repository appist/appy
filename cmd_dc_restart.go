package appy

import (
	"net/http"
)

func newDcRestartCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:restart",
		Short: "Restart services that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
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
