package cmd

import (
	"os"
	"path"

	"github.com/appist/appy/support"

	"github.com/spf13/cobra"
)

var (
	logger           *support.LoggerT
	root             *cobra.Command
	reservedCmdNames = map[string]bool{}
)

func init() {
	logger = support.Logger
	root = NewCommand()
}

// NewCommand initializes the root command instance.
func NewCommand() *cobra.Command {
	cmdName := path.Base(os.Args[0])
	if cmdName == "main" {
		wd, err := os.Getwd()
		if err != nil {
			logger.Fatal(err)
		}

		cmdName = path.Base(wd)
	}

	return &cobra.Command{
		Use:     cmdName,
		Short:   "An opinionated productive web framework that helps scaling business easier.",
		Version: support.VERSION,
	}
}

// AddCommand adds a custom command.
func AddCommand(command *cobra.Command) {
	if _, ok := reservedCmdNames[command.Name()]; ok {
		logger.Fatalf("'%s' command name is reserved, please update the command name.", command.Name())
	}

	root.AddCommand(command)
}

// Run executes the root command.
func Run() {
	root.Execute()
}
