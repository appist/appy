package cmd

import (
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbSeedCommand dump all the databases schema for the current environment (debug build only).
func NewDbSeedCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	var target string

	cmd := &Command{
		Use:   "db:schema:dump",
		Short: "Dump all the databases schema for the current environment (debug build only)",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			runDbSeedAll(config, dbManager, logger, target)
		},
	}

	cmd.Flags().StringVar(&target, "database", "", "The target database to seed")
	return cmd
}

func runDbSeedAll(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger, target string) {
	err := dbManager.ConnectAll(true)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbManager.CloseAll()

	if target != "" {
		db := dbManager.Db(target)
		if db == nil {
			logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
			os.Exit(0)
		}

		if db.Config().Replica {
			logger.Infof("Unable to run seeding on '%s' database that is a replica", target)
			os.Exit(0)
		}

		logger.Infof("Seeding '%s' database...", target)
		err := db.Seed()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Seeding '%s' database... DONE", target)

		return
	}

	for name, db := range dbManager.Dbs() {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Seeding '%s' database...", name)
		err := db.Seed()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Seeding '%s' database... DONE", name)
	}
}
