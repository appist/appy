package appy

import "os"

func newDbRollbackCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	var target string

	cmd := &Cmd{
		Use:   "db:rollback",
		Short: "Rolls back the database(default: primary, use --database for another one) to previous version for the current environment",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			CheckDbManager(config, dbManager, logger)

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			err := dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			db := dbManager.Db(target)
			if db == nil {
				logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
				os.Exit(0)
			}

			if db.config.Replica {
				logger.Infof("Unable to run rollback on '%s' database that is a replica", target)
				os.Exit(0)
			}

			logger.Infof("Rolling back '%s' database...", target)
			err = db.Rollback()
			if err != nil {
				logger.Fatal(err)
			}

			err = db.dumpSchema(target)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infof("Rolling back '%s' database... DONE", target)

			return
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to rollback")
	return cmd
}
