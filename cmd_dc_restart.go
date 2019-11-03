package appy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func newDcRestartCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:restart",
		Short: "Restart services that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = dcRestart(assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}

func dcRestart(assets http.FileSystem) error {
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

	dcRestartCmd := exec.Command("docker-compose", "-f", "-", "-p", appName, "restart")
	dcRestartCmd.Stdin = bytes.NewBuffer(data)
	dcRestartCmd.Stdout = os.Stdout
	dcRestartCmd.Stderr = os.Stderr
	dcRestartCmd.Run()

	return nil
}
