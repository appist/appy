package cmd

import (
	"github.com/appist/appy/support"
)

func newDcDownCommand(asset *support.Asset, logger *support.Logger) *Command {
	return &Command{
		Use:   "dc:down",
		Short: "Tear down the docker compose cluster",
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
