package cmd

import (
	"fmt"
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbDropCommand drop all databases for the current environment.
func NewDbDropCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	cmd := &Command{
		Use:   "db:drop",
		Short: "Drop all databases for the current environment",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsProtectedEnv(config) {
				fmt.Printf("You are attempting to run a destructive action against your database in '%s' environment.\n", config.AppyEnv)
				os.Exit(-1)
			}

			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			logger.Infof("Dropping databases from pkg/config/.env.%s...", config.AppyEnv)

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

				tmpErrs := db.Drop()
				errs = append(errs, tmpErrs...)
			}

			if len(errs) > 0 {
				for _, err := range errs {
					logger.Infof(err.Error())
				}

				os.Exit(-1)
			}
		},
	}

	return cmd
}
