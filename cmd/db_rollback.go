package cmd

import (
	"os"

	"github.com/appist/appy/core"
)

// NewDbRollbackCommand rolls back specific database(default: primary) to the last migration for the current environment.
func NewDbRollbackCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	var target string

	cmd := &AppCmd{
		Use:   "db:rollback",
		Short: "Rolls back specific database(default: primary) to the last migration for the current environment.",
		Run: func(cmd *AppCmd, args []string) {
			logger.Infof("Rolling back '%s' database from app/config/.env.%s...", target, config.AppyEnv)

			err := core.DbConnect(dbMap, true)
			if err != nil {
				logger.Fatal(err)
			}
			defer core.DbClose(dbMap)

			if _, ok := dbMap[target]; !ok {
				logger.Infof("'%s' database is not defined in app/config/.env.%s.", target, config.AppyEnv)
				os.Exit(-1)
			}

			err = dbMap[target].Rollback()
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to migrate.")
	return cmd
}
