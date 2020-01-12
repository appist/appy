package appy

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

// Command is used to build the command line interface.
type Command struct {
	*cobra.Command
	config *Config
}

// NewCommand initializes Command instance.
func NewCommand(config *Config) *Command {
	commandName := path.Base(os.Args[0])
	if commandName == "main" {
		wd, _ := os.Getwd()
		commandName = path.Base(wd)
	}

	command := &Command{
		&cobra.Command{},
		config,
	}
	command.Use = commandName
	command.Short = DESCRIPTION
	command.Version = VERSION

	return command
}
