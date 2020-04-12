package cmd

import (
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newDBCreateCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
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

			if len(dbManager.Databases()) < 1 {
				logger.Fatalf("No database is defined in '%s'", config.Path())
			}

			runDBCreateAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBCreateAll(config *support.Config, dbManager *record.Engine, logger *support.Logger) {
	for name, db := range dbManager.Databases() {
		dbConfig := db.Config()
		if dbConfig.Replica {
			continue
		}

		logger.Infof("Creating '%s' database...", name)

		err := db.CreateDB(dbConfig.Database)
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
