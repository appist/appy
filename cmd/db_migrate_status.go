package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbMigrateStatusCommand displays status of migrations for all databases in the current environment.
func NewDbMigrateStatusCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:migrate:status",
		Short: "Displays status of migrations for all databases in the current environment.",
		Run: func(cmd *AppCmd, args []string) {
		},
	}

	return cmd
}
