package appy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

type (
	// Cmd defines what a command line can do.
	Cmd = cobra.Command
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

	appName string
)

// NewCmd initializes the root command instance.
func NewCmd() *Cmd {
	appName = path.Base(os.Args[0])
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

func runDockerCompose(action string, assets http.FileSystem) error {
	var (
		data []byte
		err  error
	)
	dcPath := _ssrPaths["docker"] + "/docker-compose.yml"

	if Build == DebugBuild {
		data, err = ioutil.ReadFile(dcPath)
		if err != nil {
			return err
		}
	} else {
		file, err := assets.Open(_ssrPaths["root"] + "/" + dcPath)
		if err != nil {
			return err
		}

		data, err = ioutil.ReadAll(file)
		if err != nil {
			return err
		}
	}

	var cmd *exec.Cmd
	switch action {
	case "down":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", appName, action, "--remove-orphans")
	case "up":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", appName, action, "-d")
	case "restart":
		cmd = exec.Command("docker-compose", "-f", "-", "-p", appName, action)
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
