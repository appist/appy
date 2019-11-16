package cmd

import (
	"net/http"

	appysupport "github.com/appist/appy/internal/support"
)

// NewTeardownCommand run dc:down to teardown the docker-compose cluster.
func NewTeardownCommand(logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "teardown",
		Short: "Run dc:down to teardown the docker-compose cluster",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("down", assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}
