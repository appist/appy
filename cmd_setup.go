//+build !test

package appy

import "time"

func newSetupCommand(asset *Asset, config *Config, dbManager *DBManager, logger *Logger) *Command {
	return &Command{
		Use:   "setup",
		Short: "Run dc:up/db:create/db:schema:load/db:seed to setup the datastore with seed data",
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

			err = runDockerCompose("up", asset)
			if err != nil {
				logger.Fatal(err)
			}

			time.Sleep(10 * time.Second)
			runDBCreateAll(config, dbManager, logger)
			runDBSchemaLoad(config, dbManager, logger)
			runDBSeedAll(config, dbManager, logger)
		},
	}
}
