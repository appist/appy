package cmd

import (
	"os"
	"path"

	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

type (
	// Command defines what a command line can do.
	Command = cobra.Command
)

// NewCommand initializes the root command instance.
func NewCommand() *Command {
	return &Command{
		Use:     getCommandName(),
		Short:   support.DESCRIPTION,
		Version: support.VERSION,
	}
}

func getCommandName() string {
	name := path.Base(os.Args[0])
	if name == "main" {
		wd, _ := os.Getwd()
		name = path.Base(wd)
	}

	return name
}
