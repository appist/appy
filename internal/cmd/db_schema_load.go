package cmd

import (
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbSchemaLoadCommand load all the databases schema for the current environment (debug build only).
func NewDbSchemaLoadCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	cmd := &Command{
		Use:   "db:schema:load",
		Short: "Load all the databases schema for the current environment (debug build only)",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			logger.SetDbLogging(false)
			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			runDbSchemaLoad(config, dbManager, logger)
		},
	}

	return cmd
}

func runDbSchemaLoad(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) {
	logger.SetDbLogging(false)
	err := dbManager.ConnectAll(true)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbManager.CloseAll()

	for name, db := range dbManager.Dbs() {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Loading schema for '%s' database...", name)
		_, err := db.Handle().Exec(db.Schema())
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Loading schema for '%s' database... DONE", name)
	}
}
