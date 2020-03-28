//+build !test

package appy

import "strings"

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
	for name, db := range dbManager.databases {
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
