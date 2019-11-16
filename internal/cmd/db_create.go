package cmd

import (
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbCreateCommand create all databases for the current environment.
func NewDbCreateCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	cmd := &Command{
		Use:   "db:create",
		Short: "Create all databases for the current environment",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			runDbCreateAll(config, dbManager, logger)
		},
	}

	return cmd
}

func runDbCreateAll(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) {
	logger.Infof("Creating databases from pkg/config/.env.%s...", config.AppyEnv)

	err := dbManager.ConnectAll(false)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbManager.CloseAll()

	var errs []error
	for _, db := range dbManager.Dbs() {
		if db.Config().Replica {
			continue
		}

		tmpErrs := db.Create()
		errs = append(errs, tmpErrs...)
	}

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Infof(err.Error())
		}

		os.Exit(-1)
	}
}
