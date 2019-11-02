package appy

import (
	"errors"
	"os"
)

func newGMigrationCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	var (
		tx     bool
		target string
	)

	cmd := &Cmd{
		Use:   "g:migration",
		Short: "Generates a database migration file(default: primary, use --database to specify another 1) for the current environment",
		Args: func(cmd *Cmd, args []string) error {
			if len(args) < 1 || !IsPascalCase(args[0]) {
				return errors.New("please provide migration name in pascal case, e.g. CreateUsers")
			}

			return nil
		},
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			CheckDbManager(config, dbManager, logger)

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to generate the migration file for.")
	cmd.Flags().BoolVar(&tx, "tx", true, "Indicate whether or not to run the migration in a transaction.")
	return cmd
}
