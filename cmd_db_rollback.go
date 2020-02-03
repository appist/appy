//+build !test

package appy

func newDBRollbackCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	var target string

	cmd := &Command{
		Use:   "db:rollback",
		Short: "Rollback the database(default: primary, use --database to specify the target database) to previous version for the current environment",
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

			db := dbManager.DB(target)
			if db == nil {
				logger.Fatalf("No database called '%s' defined in 'configs/.env.%s'", target, config.AppyEnv)
			}

			if db.Config().Replica {
				logger.Fatalf("Unable to run rollback on '%s' database that is a replica", target)
			}

			logger.Infof("Rolling back '%s' database...", target)

			err := db.Connect()
			if err != nil {
				logger.Fatal(err)
			}
			defer db.Close()

			err = db.Rollback()
			if err != nil {
				logger.Fatal(err)
			}

			if IsDebugBuild() {
				err = db.DumpSchema(target)
				if err != nil {
					logger.Fatal(err)
				}
			}

			logger.Infof("Rolling back '%s' database... DONE", target)
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to rollback")
	return cmd
}
