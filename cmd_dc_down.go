package appy

import (
	"net/http"
)

func newDcDownCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:down",
		Short: "Stop and remove containers, networks, images, and volumes that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
			runDcDown(logger, assets)
		},
	}

	return cmd
}

func runDcDown(logger *Logger, assets http.FileSystem) {
	err := checkDocker()
	if err != nil {
		logger.Fatal(err)
	}

	err = runDockerCompose("down", assets)
	if err != nil {
		logger.Fatal(err)
	}
}
