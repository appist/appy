package appy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func newDcUpCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:up",
		Short: "Create and start containers that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
			_, err := exec.LookPath("docker-compose")
			if err != nil {
				logger.Fatal(err)
			}

			err = dcUp(assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}

func dcUp(assets http.FileSystem) error {
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

	dcUpCmd := exec.Command("docker-compose", "-f", "-", "-p", appName, "up", "-d")
	dcUpCmd.Stdin = bytes.NewBuffer(data)
	dcUpCmd.Stdout = os.Stdout
	dcUpCmd.Stderr = os.Stderr
	dcUpCmd.Run()

	return nil
}
