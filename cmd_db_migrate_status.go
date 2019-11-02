package appy

import (
	"fmt"
	"os"

	"github.com/bndr/gotabulate"
)

func newDbMigrateStatusCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	cmd := &Cmd{
		Use:   "db:migrate:status",
		Short: "Migrates the database(default: all, use --database to specify just 1) for the current environment",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)
			CheckDbManager(config, dbManager, logger)
			logger.SetDbLogging(false)

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			err := dbManager.ConnectAll(true)
			if err != nil {
				logger.Fatal(err)
			}
			defer dbManager.CloseAll()

			for name, db := range dbManager.dbs {
				if db.config.Replica {
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
