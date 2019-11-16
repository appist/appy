package cmd

import (
	"fmt"
	"os"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
	"github.com/bndr/gotabulate"
)

// NewDbMigrateStatusCommand list all the database migration status(default: all, use --database to specify just 1) for the current environment.
func NewDbMigrateStatusCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) *Command {
	cmd := &Command{
		Use:   "db:migrate:status",
		Short: "List all the database migration status(default: all, use --database to specify just 1) for the current environment",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			logger.SetDbLogging(false)
			if len(dbManager.Dbs()) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			err := dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			for name, db := range dbManager.Dbs() {
				if db.Config().Replica {
					continue
				}

				fmt.Println()
				fmt.Printf("database: %s\n", name)
				fmt.Println()

				status, err := db.MigrateStatus()
				if err != nil {
					logger.Fatal(err)
				}

				table := gotabulate.Create(status)
				table.SetAlign("left")
				table.SetHeaders([]string{"Status", "Migration ID", "Migration File"})
				fmt.Println(table.Render("simple"))
			}
		},
	}

	return cmd
}
