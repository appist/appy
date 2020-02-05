//+build !test

package appy

func newTeardownCommand(asset *Asset, config *Config, dbManager *DBManager, logger *Logger) *Command {
	return &Command{
		Use:   "teardown",
		Short: "Run dc:down to teardown the Docker cluster",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

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
