package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbMigrateCommand migrates the database.
func NewDbMigrateCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:migrate",
		Short: "Migrates the database.",
		Run: func(cmd *AppCmd, args []string) {
		},
	}

	return cmd
}
