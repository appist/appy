//+build !test

package appy

func newDBSchemaLoadCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	cmd := &Command{
		Use:   "db:schema:load",
		Short: "Load all the databases schema for the current environment",
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

			runDBSchemaLoad(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBSchemaLoad(config *Config, dbManager *DBManager, logger *Logger) {
	logger.SetDbLogging(false)

	for name, db := range dbManager.databases {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Loading schema for '%s' database...", name)

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		_, err = db.Exec(db.Schema())
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Loading schema for '%s' database... DONE", name)
	}
}
