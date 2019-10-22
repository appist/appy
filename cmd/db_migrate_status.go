package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbMigrateStatusCommand displays status of migrations.
func NewDbMigrateStatusCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:migrate",
		Short: "Displays status of migrations.",
		Run: func(cmd *AppCmd, args []string) {
		},
	}

	return cmd
}
