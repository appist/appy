package cmd

import (
	"os/exec"

	"github.com/appist/appy/core"
)

// NewDbSchemaDumpCommand dumps all the databases schema for the current environment.
func NewDbSchemaDumpCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:schema:dump",
		Short: "Dumps all the databases schema for the current environment.",
		Run: func(cmd *AppCmd, args []string) {
			core.SetDbLogging(false)
			_, err := exec.LookPath("pg_dump")
			if err != nil {
				logger.Fatal(err)
			}

			err = core.DbConnect(dbMap, true)
			if err != nil {
				logger.Fatal(err)
			}
			defer core.DbClose(dbMap)

			for name, db := range dbMap {
				if db.Config.Replica {
					continue
				}

				err := dumpSchema(name, db)
				if err != nil {
					logger.Fatal(err)
				}
			}
		},
	}

	return cmd
}
