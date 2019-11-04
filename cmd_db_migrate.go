package appy

import "os"

func newDbMigrateCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	var target string

	cmd := &Cmd{
		Use:   "db:migrate",
		Short: "Migrate the database(default: all, use --database to specify just 1) for the current environment",
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(config, logger) || IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			err := dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			if target != "" {
				db := dbManager.Db(target)
				if db == nil {
					logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
					os.Exit(0)
				}

				if db.config.Replica {
					logger.Infof("Unable to run migration on '%s' database that is a replica", target)
					os.Exit(0)
				}

				logger.Infof("Migrating '%s' database...", target)
				err := db.Migrate()
				if err != nil {
					logger.Fatal(err)
				}

				err = db.dumpSchema(target)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Migrating '%s' database... DONE", target)

				return
			}

			for name, db := range dbManager.dbs {
				if db.config.Replica {
					continue
				}

				logger.Infof("Migrating '%s' database...", name)
				err := db.Migrate()
				if err != nil {
					logger.Fatal(err)
				}

				err = db.dumpSchema(name)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Migrating '%s' database... DONE", name)
			}
		},
	}

	cmd.Flags().StringVar(&target, "database", "", "The target database to migrate")
	return cmd
}
