//+build !test

package appy

import (
	"fmt"

	"github.com/bndr/gotabulate"
)

func newDBMigrateStatusCommand(config *Config, dbManager *DBManager, logger *Logger) *Command {
	return &Command{
		Use:   "db:migrate:status",
		Short: "List all the database migration status(default: all, use --database to specify the target database) for the current environment",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			if len(dbManager.databases) < 1 {
				logger.Fatalf("No database is defined in 'configs/.env.%s'", config.AppyEnv)
			}

			for name, db := range dbManager.databases {
				if db.Config().Replica {
					continue
				}

				err := db.Connect()
				if err != nil {
					logger.Fatal(err)
				}
				defer db.Close()

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
}
