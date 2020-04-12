package cmd

import (
	"strings"

	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newDBSchemaLoadCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
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

			if len(dbManager.Databases()) < 1 {
				logger.Fatalf("No database is defined in '%s'", config.Path())
			}

			runDBSchemaLoad(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBSchemaLoad(config *support.Config, dbManager *record.Engine, logger *support.Logger) {
	for name, db := range dbManager.Databases() {
		dbConfig := db.Config()
		if dbConfig.Replica {
			continue
		}

		query := db.Schema()
		if dbConfig.Adapter == "mysql" {
			query = "\nSET FOREIGN_KEY_CHECKS = 0;\n" + query + "\n\nSET FOREIGN_KEY_CHECKS = 1;"

			if !strings.Contains(dbConfig.URI, "multiStatements=true") {
				if !strings.Contains(dbConfig.URI, "?") {
					dbConfig.URI += "?"
				}

				dbConfig.URI += "multiStatements=true"
			}
		}

		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
		defer db.Close()

		logger.Infof("Loading schema for '%s' database...", name)

		_, err = db.Exec(query)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Infof("Loading schema for '%s' database... DONE", name)
	}
}
