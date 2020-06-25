package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/worker"
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

// NewCommand initializes Command instance without built-in commands.
func NewCommand() *Command {
	return &Command{
		Use:     getCommandName(),
		Short:   support.DESCRIPTION,
		Version: support.VERSION,
	}
}

// NewAppCommand initializes Command instance without built-in commands.
func NewAppCommand(asset *support.Asset, config *support.Config, dbManager *record.Engine, logger *support.Logger, server *pack.Server, worker *worker.Engine) *Command {
	cmd := NewCommand()
	cmd.AddCommand(newDBCreateCommand(config, dbManager, logger))
	cmd.AddCommand(newDBDropCommand(config, dbManager, logger))
	cmd.AddCommand(newDBMigrateCommand(config, dbManager, logger))
	cmd.AddCommand(newDBMigrateStatusCommand(config, dbManager, logger))
	cmd.AddCommand(newDBRollbackCommand(config, dbManager, logger))
	cmd.AddCommand(newDBSchemaLoadCommand(config, dbManager, logger))
	cmd.AddCommand(newDBSeedCommand(config, dbManager, logger))
	cmd.AddCommand(newDcDownCommand(asset, logger))
	cmd.AddCommand(newDcRestartCommand(asset, logger))
	cmd.AddCommand(newDcUpCommand(asset, logger))
	cmd.AddCommand(newMiddlewareCommand(config, logger, server))
	cmd.AddCommand(newRoutesCommand(config, logger, server))
	cmd.AddCommand(newSecretCommand(logger))
	cmd.AddCommand(newServeCommand(dbManager, logger, server))
	cmd.AddCommand(newSetupCommand(asset, config, dbManager, logger))
	cmd.AddCommand(newSSLSetupCommand(logger, server))
	cmd.AddCommand(newSSLTearDownCommand(logger, server))
	cmd.AddCommand(newTearDownCommand(asset, logger))
	cmd.AddCommand(newWorkCommand(config, dbManager, logger, worker))

	if support.IsDebugBuild() {
		cmd.AddCommand(newBuildCommand(asset, logger, server))
		cmd.AddCommand(newConfigDecCommand(config, logger))
		cmd.AddCommand(newConfigEncCommand(config, logger))
		cmd.AddCommand(newDBSchemaDumpCommand(config, dbManager, logger))
		cmd.AddCommand(newGenMigrationCommand(config, dbManager, logger))
		cmd.AddCommand(newSecretRotateCommand(asset, config, logger))
		cmd.AddCommand(newStartCommand(logger, server))
	}

	return cmd
}

func getCommandName() string {
	name := path.Base(os.Args[0])
	if name == "main" {
		wd, _ := os.Getwd()
		name = path.Base(wd)
	}

	return name
}

func checkDocker() error {
	binaries := []string{"docker", "docker-compose"}

	for _, binary := range binaries {
		_, err := exec.LookPath(binary)
		if err != nil {
			return err
		}
	}

	return nil
}

func runDockerCompose(action string, asset *support.Asset) error {
	var (
		data []byte
		err  error
	)
	dcPath := asset.Layout().Docker() + "/docker-compose.yml"

	if support.IsDebugBuild() {
		data, err = ioutil.ReadFile(dcPath)
		if err != nil {
			return err
		}
	} else {
		file, err := asset.Open("/" + dcPath)
		if err != nil {
			return err
		}

		data, err = ioutil.ReadAll(file)
		if err != nil {
			return err
		}
	}

	var cmd *exec.Cmd
	clusterName := getCommandName()
	switch action {
	case "down":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", clusterName, action, "--remove-orphans")
	case "up":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", clusterName, action, "-d")
	case "restart":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", clusterName, action)
	}

	cmd.Stdin = bytes.NewBuffer(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
