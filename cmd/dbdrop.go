package cmd

import (
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newDBDropCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
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

			if len(dbManager.Databases()) < 1 {
				logger.Fatalf("No database is defined in '%s'", config.Path())
			}

			runDBDropAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBDropAll(config *support.Config, dbManager *record.Engine, logger *support.Logger) {
	for name, db := range dbManager.Databases() {
		dbConfig := db.Config()
		if dbConfig.Replica {
			continue
		}

		logger.Infof("Dropping '%s' database...", name)

		err := db.DropDB(dbConfig.Database)
		if err != nil {
			logger.Fatal(err)
		}

		err = db.DropDB(dbConfig.Database + "_test")
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Dropping '%s' database... DONE", name)
	}
}
