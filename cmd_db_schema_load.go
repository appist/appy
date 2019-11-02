package appy

import (
	"os"
)

func newDbSchemaLoadCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	cmd := &Cmd{
		Use:   "db:schema:load",
		Short: "Loads all the databases schema for the current environment (debug build only)",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			CheckDbManager(config, dbManager, logger)

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			logger.SetDbLogging(false)
			err := dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			for name, db := range dbManager.dbs {
				if db.config.Replica {
					continue
				}

				logger.Infof("Loading schema for '%s' database...", name)
				_, err := db.handle.Exec(db.schema)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Loading schema for '%s' database... DONE", name)
			}
		},
	}

	return cmd
}
