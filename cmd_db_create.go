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
		if db.Config().Replica {
			continue
		}

		logger.Infof("Creating '%s' database...", name)

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		errs := db.Create()
		if errs != nil {
			logger.Fatal(errs[0])
		}

		logger.Infof("Creating '%s' database... DONE", name)
	}
}
