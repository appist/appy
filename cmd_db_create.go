//+build !test

package appy

func newDBCreateCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	cmd := &Command{
		Use:   "db:create",
		Short: "Create all databases for the current environment",
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

			runDBCreateAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBCreateAll(config *Config, dbManager *DBManager, logger *Logger) {
	for name, db := range dbManager.databases {
		dbConfig := db.Config()
		if dbConfig.Replica {
			continue
		}

		err := db.ConnectDB(dbConfig.Adapter)
		defer db.Close()

		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Creating '%s' database...", name)

		err = db.CreateDB(dbConfig.Database)
		if err != nil {
			logger.Fatal(err)
		}

		err = db.CreateDB(dbConfig.Database + "_test")
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Creating '%s' database... DONE", name)
	}
}
