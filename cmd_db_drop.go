//+build !test

package appy

func newDBDropCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	cmd := &Command{
		Use:   "db:drop",
		Short: "Drop all databases for the current environment",
		Run: func(cmd *Command, args []string) {
			if config.IsProtectedEnv() {
				logger.Fatal("You are attempting to run a destructive action against your database in '%s' environment.\n", config.AppyEnv)
			}

			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			if len(dbManager.databases) < 1 {
				logger.Fatalf("No database is defined in 'configs/.env.%s'", config.AppyEnv)
			}

			runDBDropAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBDropAll(config *Config, dbManager *DBManager, logger *Logger) {
	for name, db := range dbManager.databases {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Dropping '%s' database...", name)

		targetDB := db.config.Database
		db.config.Database = "postgres"
		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		db.config.Database = targetDB
		errs := db.Drop()
		if errs != nil {
			logger.Fatal(errs[0])
		}

		logger.Infof("Dropping '%s' database... DONE", name)
	}
}
