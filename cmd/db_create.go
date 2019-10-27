package cmd

import (
	"os"

	"github.com/appist/appy/core"
)

// NewDbCreateCommand creates all databases for the current environment.
func NewDbCreateCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:create",
		Short: "Creates all databases for the current environment.",
		Run: func(cmd *AppCmd, args []string) {
			logger.Infof("Creating databases from app/config/.env.%s...", config.AppyEnv)

			err := core.DbConnect(dbMap, false)
			if err != nil {
				logger.Fatal(err)
			}
			defer core.DbClose(dbMap)

			if len(dbMap) < 1 {
				logger.Infof("No database is defined in app/config/.env.%s.", config.AppyEnv)
				os.Exit(-1)
			}

			var errs []string
			for _, db := range dbMap {
				if db.Config.Replica {
					continue
				}

				tmpErrs := dbCreate(db)
				errs = append(errs, tmpErrs...)
			}

			if len(errs) > 0 {
				for _, err := range errs {
					logger.Infof(err)
				}

				os.Exit(-1)
			}
		},
	}

	return cmd
}
