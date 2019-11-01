package appy

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

type (
	// Cmd defines what a command line can do.
	Cmd = cobra.Command
)

// NewCmd initializes the root command instance.
func NewCmd() *Cmd {
	appName := path.Base(os.Args[0])
	if appName == "main" {
		wd, _ := os.Getwd()
		appName = path.Base(wd)
	}

	return &Cmd{
		Use:     appName,
		Short:   _description,
		Version: VERSION,
	}
}
