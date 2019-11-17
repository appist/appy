package cmd

import (
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewDbMigrateCommand migrate the database(default: all, use --database to specify just 1) for the current environment.
func NewDbMigrateCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	var target string

	cmd := &Command{
		Use:   "db:migrate",
		Short: "Migrate the database(default: all, use --database to specify just 1) for the current environment",
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

			if target != "" {
				db := dbManager.Db(target)
				if db == nil {
					logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
					os.Exit(0)
				}

				if db.Config().Replica {
					logger.Infof("Unable to run migration on '%s' database that is a replica", target)
					os.Exit(0)
				}

				logger.Infof("Migrating '%s' database...", target)
				err := db.Migrate()
				if err != nil {
					logger.Fatal(err)
				}

				err = db.DumpSchema(target)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Migrating '%s' database... DONE", target)

				return
			}

			for name, db := range dbManager.Dbs() {
				if db.Config().Replica {
					continue
				}

				logger.Infof("Migrating '%s' database...", name)
				err := db.Migrate()
				if err != nil {
					logger.Fatal(err)
				}

				err = db.DumpSchema(name)
				if err != nil {
					logger.Fatal(err)
				}
				logger.Infof("Migrating '%s' database... DONE", name)
			}
		},
	}

	cmd.Flags().StringVar(&target, "database", "", "The target database to migrate")
	return cmd
}