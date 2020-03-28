//+build !test

package appy

import (
	"strings"
)

func newDBDropCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
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

			if len(dbManager.databases) < 1 {
				logger.Fatalf("No database is defined in 'configs/.env.%s'", config.AppyEnv)
			}

			runDBDropAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDBDropAll(config *Config, dbManager *DBManager, logger *Logger) {
	for name, db := range dbManager.databases {
		dbConfig := db.Config()
		if dbConfig.Replica {
			continue
		}

		if dbConfig.Adapter != "sqlite3" {
			err := db.ConnectDB(dbConfig.Adapter)
			defer db.Close()

			if err != nil {
				logger.Fatal(err)
			}
		}

		logger.Infof("Dropping '%s' database...", name)

		switch dbConfig.Adapter {
		case "mysql", "postgres":
			err := db.DropDB(dbConfig.Database)
			if err != nil {
				logger.Fatal(err)
			}

			err = db.DropDB(dbConfig.Database + "_test")
			if err != nil {
				logger.Fatal(err)
			}
		case "sqlite3":
			err := db.DropDB(dbConfig.URI)
			if err != nil {
				logger.Fatal(err)
			}

			err = db.DropDB(strings.ReplaceAll(dbConfig.URI, ".sqlite3", "_test.sqlite3"))
			if err != nil {
				logger.Fatal(err)
			}
		}

		logger.Infof("Dropping '%s' database... DONE", name)
	}
}
