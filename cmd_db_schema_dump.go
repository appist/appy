package appy

import (
	"os"
	"os/exec"
)

func newDbSchemaDumpCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	cmd := &Cmd{
		Use:   "db:schema:dump",
		Short: "Dump all the databases schema for the current environment (debug build only)",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			CheckDbManager(config, dbManager, logger)

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			logger.SetDbLogging(false)
			_, err := exec.LookPath("pg_dump")
			if err != nil {
				logger.Fatal(err)
			}

			err = dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			for name, db := range dbManager.dbs {
				if db.config.Replica {
					continue
				}

				logger.Infof("Dumping schema for '%s' database...", name)
				err := db.dumpSchema(name)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Dumping schema for '%s' database... DONE", name)
			}
		},
	}

	return cmd
}
