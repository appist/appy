package cmd

import (
	"errors"
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewGenMigrationCommand generate database migration file(default: primary, use --database to specify another 1) for the current environment.
func NewGenMigrationCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	var (
		tx     bool
		target string
	)

	cmd := &Command{
		Use:   "gen:migration [name]",
		Short: "Generate database migration file(default: primary, use --database to specify another 1) for the current environment",
		Args: func(cmd *Command, args []string) error {
			if len(args) < 1 || !appysupport.IsPascalCase(args[0]) {
				return errors.New("please provide migration name in pascal case, e.g. CreateUsers")
			}

			return nil
		},
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			db := dbManager.Db(target)
			if db == nil {
				logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
				os.Exit(0)
			}

			db.GenerateMigration(args[0], target, tx)
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to generate the migration file for")
	cmd.Flags().BoolVar(&tx, "tx", true, "Indicate whether or not to run the migration in a transaction")
	return cmd
}
