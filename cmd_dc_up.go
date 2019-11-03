package appy

import (
	"net/http"
)

func newDcUpCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:up",
		Short: "Create and start containers that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
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
