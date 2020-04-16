package cmd

import (
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newDBMigrateCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
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

			if len(dbManager.Databases()) < 1 {
				logger.Fatalf("No database is defined in '%s'", config.Path())
			}

			if target != "" {
				db := dbManager.DB(target)
				if db == nil {
					logger.Fatalf("No database called '%s' defined in '%s'", target, config.Path())
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

				if support.IsDebugBuild() {
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

func runDBMigrateAll(config *support.Config, dbManager *record.Engine, logger *support.Logger) {
	for name, db := range dbManager.Databases() {
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

		if support.IsDebugBuild() {
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
