//+build !test

package appy

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

// Command is used to build the command line interface.
type Command = cobra.Command

var (
	// ExactArgs returns an error if there are not exactly n args.
	ExactArgs = cobra.ExactArgs

	// ExactValidArgs returns an error if
	// there are not exactly N positional args OR
	// there are any positional args that are not in the `ValidArgs` field of `Command`
	ExactValidArgs = cobra.ExactValidArgs

	// MinimumNArgs returns an error if there is not at least N args.
	MinimumNArgs = cobra.MinimumNArgs

	// MaximumNArgs returns an error if there are more than N args.
	MaximumNArgs = cobra.MaximumNArgs

	// NoArgs returns an error if any args are included.
	NoArgs = cobra.NoArgs

	// OnlyValidArgs returns an error if any args are not in the list of ValidArgs.
	OnlyValidArgs = cobra.OnlyValidArgs

	// RangeArgs returns an error if the number of args is not within the expected range.
	RangeArgs = cobra.RangeArgs
)

// NewRootCommand initializes Command instance.
func NewRootCommand() *Command {
	commandName := path.Base(os.Args[0])
	if commandName == "main" {
		wd, _ := os.Getwd()
		commandName = path.Base(wd)
	}

	command := &Command{
		Use:     commandName,
		Short:   DESCRIPTION,
		Version: VERSION,
	}

	return command
}
