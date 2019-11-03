package appy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func newDcDownCommand(logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "dc:down",
		Short: "Stop and remove containers, networks, images, and volumes that are defined in .docker/docker-compose.yml",
		Run: func(cmd *Cmd, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = dcDown(assets)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	return cmd
}

func dcDown(assets http.FileSystem) error {
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

	dcDownCmd := exec.Command("docker-compose", "-f", "-", "-p", appName, "down", "--remove-orphans")
	dcDownCmd.Stdin = bytes.NewBuffer(data)
	dcDownCmd.Stdout = os.Stdout
	dcDownCmd.Stderr = os.Stderr
	dcDownCmd.Run()

	return nil
}
