//+build !test

package appy

func newDBSeedCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
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

			if len(dbManager.databases) < 1 {
				logger.Fatalf("No database is defined in 'configs/.env.%s'", config.AppyEnv)
			}

			runDBSeedAll(config, dbManager, logger)
		},
	}
}

func runDBSeedAll(config *Config, dbManager *DBManager, logger *Logger) {
	for name, db := range dbManager.databases {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Seeding '%s' database...", name)

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		err = db.Seed()
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Seeding '%s' database... DONE", name)
	}
}
