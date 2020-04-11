package cmd

import (
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newDBSeedCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
	return &Command{
		Use:   "db:seed",
		Short: "Seed all databases for the current environment",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			if len(dbManager.Databases()) < 1 {
				logger.Fatalf("No database is defined in 'configs/.env.%s'", config.AppyEnv)
			}

			runDBSeedAll(config, dbManager, logger)
		},
	}
}

func runDBSeedAll(config *support.Config, dbManager *record.Engine, logger *support.Logger) {
	for name, db := range dbManager.Databases() {
		if db.Config().Replica {
			continue
		}

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		logger.Infof("Seeding '%s' database...", name)

		err = db.Seed()
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Seeding '%s' database... DONE", name)
	}
}
