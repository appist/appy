package cmd

import (
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbRollbackCommand rollback the database(default: primary, use --database to specify another 1) to previous version for the current environment.
func NewDbRollbackCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	var target string

	cmd := &Command{
		Use:   "db:rollback",
		Short: "Rollback the database(default: primary, use --database to specify another 1) to previous version for the current environment",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
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

			if db.Config().Replica {
				logger.Infof("Unable to run rollback on '%s' database that is a replica", target)
				os.Exit(0)
			}

			logger.Infof("Rolling back '%s' database...", target)
			err = db.Rollback()
			if err != nil {
				logger.Fatal(err)
			}

			err = db.DumpSchema(target)
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infof("Rolling back '%s' database... DONE", target)
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to rollback")
	return cmd
}
