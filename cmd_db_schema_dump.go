//+build !test

package appy

import (
	"os/exec"
)

func newDBSchemaDumpCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	cmd := &Command{
		Use:   "db:schema:dump",
		Short: "Dump all the databases schema for the current environment (only available in debug build)",
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

			_, err := exec.LookPath("pg_dump")
			if err != nil {
				logger.Fatal(err)
			}

			for name, db := range dbManager.databases {
				if db.Config().Replica {
					continue
				}

				err := db.Connect()
				if err != nil {
					logger.Fatal(err)
				}
				defer db.Close()

				logger.Infof("Dumping schema for '%s' database...", name)

				err = db.DumpSchema(name)
				if err != nil {
					logger.Fatal(err)
				}

				logger.Infof("Dumping schema for '%s' database... DONE", name)
			}
		},
	}

	return cmd
}
