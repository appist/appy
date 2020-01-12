//+build !test

package appy

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

// Command is used to build the command line interface.
type Command = cobra.Command

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
