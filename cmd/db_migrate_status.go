package cmd

import (
	"fmt"

	"github.com/appist/appy/core"
	"github.com/bndr/gotabulate"
)

// NewDbMigrateStatusCommand displays status of migrations for all databases in the current environment.
func NewDbMigrateStatusCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:migrate:status",
		Short: "Displays status of migrations for all databases in the current environment.",
		Run: func(cmd *AppCmd, args []string) {
			core.SetDbLogging(false)
			err := core.DbConnect(dbMap, true)
			if err != nil {
				logger.Fatal(err)
			}
			defer core.DbClose(dbMap)

			for name, db := range dbMap {
				if db.Config.Replica {
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
