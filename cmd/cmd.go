package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

type (
	// Command defines what a command line can do.
	Command = cobra.Command
)

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

func runDockerCompose(action string, assets *support.Assets) error {
	var (
		data []byte
		err  error
	)
	dcPath := assets.Layout()["docker"] + "/docker-compose.yml"

	if support.IsDebugBuild() {
		data, err = ioutil.ReadFile(dcPath)
		if err != nil {
			return err
		}
	} else {
		file, err := assets.Open(assets.SSRRelease() + "/" + dcPath)
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
