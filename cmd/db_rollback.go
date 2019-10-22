package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbRollbackCommand rolls the schema back to the previous version.
func NewDbRollbackCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:rollback",
		Short: "Roll the schema back to the previous version.",
		Run: func(cmd *AppCmd, args []string) {
		},
	}

	return cmd
}
