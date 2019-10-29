package cmd

import (
	"github.com/appist/appy/core"
)

// NewDbSeedCommand loads the seed data for the current environment.
func NewDbSeedCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:seed",
		Short: "Loads the seed data for the current environment.",
		Run: func(cmd *AppCmd, args []string) {
		},
	}

	return cmd
}
