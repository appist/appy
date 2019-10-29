package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbSchemaLoadCommand loads all schema.go into the database for the current environment.
func NewDbSchemaLoadCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:schema:load",
		Short: "Loads all schema.go into the database for the current environment.",
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

				logger.Infof("Loading db/migrations/%s/schema.go into the database...", name)
				_, err := db.Handler.Exec(db.Schema)
				if err != nil {
					logger.Fatal(err)
				}
			}
		},
	}

	return cmd
}
