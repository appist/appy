//+build !test

package appy

import (
	"errors"
)

func newGenMigrationCommand(config *Config, dbManager *DBManager, logger *Logger, support Supporter) *Command {
	var (
		tx     bool
		target string
	)

	cmd := &Command{
		Use:   "gen:migration [name]",
		Short: "Generate database migration file(default: primary, use --database to specify the target database) for the current environment (only available in debug build)",
		Args: func(cmd *Command, args []string) error {
			if len(args) < 1 || !support.IsPascalCase(args[0]) {
				return errors.New("please provide migration name in pascal case, e.g. CreateUsers")
			}

			return nil
		},
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			db := dbManager.DB(target)
			if db == nil {
				logger.Fatalf("No database called '%s' defined in 'configs/.env.%s'", target, config.AppyEnv)
			}

			err := db.GenerateMigration(args[0], target, tx)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to generate the migration file for")
	cmd.Flags().BoolVar(&tx, "tx", true, "Indicate whether or not to run the migration in a transaction")
	return cmd
}
