//+build !test

package appy

func newDBMigrateCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	var target string

	cmd := &Command{
		Use:   "db:migrate",
		Short: "Migrate the database(default: all, use --database to specify the target database) for the current environment",
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

			if target != "" {
				db := dbManager.DB(target)
				if db == nil {
					logger.Fatalf("No database called '%s' defined in 'configs/.env.%s'", target, config.AppyEnv)
				}

				if db.Config().Replica {
					logger.Fatalf("Unable to run migration on '%s' database that is a replica", target)
				}

				logger.Infof("Migrating '%s' database...", target)

				err := db.Connect()
				if err != nil {
					logger.Fatal(err)
				}
				defer db.Close()

				err = db.Migrate()
				if err != nil {
					logger.Fatal(err)
				}

				logger.Infof("Migrating '%s' database... DONE", target)

				if IsDebugBuild() {
					logger.Infof("")
					logger.Infof("Dumping schema for '%s' database...", target)

					err = db.DumpSchema(target)
					if err != nil {
						logger.Fatal(err)
					}

					logger.Infof("Dumping schema for '%s' database... DONE", target)
				}

				return
			}

			runDBMigrateAll(config, dbManager, logger)
		},
	}

	cmd.Flags().StringVar(&target, "database", "", "The target database to migrate")
	return cmd
}

func runDBMigrateAll(config *Config, dbManager *DBManager, logger *Logger) {
	for name, db := range dbManager.databases {
		if db.Config().Replica {
			continue
		}

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		logger.Infof("Migrating '%s' database...", name)

		err = db.Migrate()
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Migrating '%s' database... DONE", name)

		if IsDebugBuild() {
			logger.Infof("")
			logger.Infof("Dumping schema for '%s' database...", name)

			err := db.DumpSchema(name)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infof("Dumping schema for '%s' database... DONE", name)
		}
	}
}
