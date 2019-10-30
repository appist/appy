package cmd

import (
	"errors"

	"github.com/appist/appy/core"
	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

// NewGenModelCommand generates a model file.
func NewGenModelCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "gen:model NAME",
		Short: "Generates a model file.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || !support.IsPascalCase(args[0]) {
				return errors.New("please provide model name in pascal case, e.g. User or UserProfile")
			}

			return nil
		},
		Run: func(cmd *AppCmd, args []string) {

		},
	}

	return cmd
}
