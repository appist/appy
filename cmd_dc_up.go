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

			var data []byte
			dcPath := ".docker/docker-compose.yml"

			if Build == DebugBuild {
				data, err = ioutil.ReadFile(dcPath)
				if err != nil {
					logger.Fatal(err)
				}
			} else {
				file, err := assets.Open(_ssrPaths["docker"] + "/" + dcPath)
				if err != nil {
					logger.Fatal(err)
				}

				data, err = ioutil.ReadAll(file)
				if err != nil {
					logger.Fatal(err)
				}
			}

			dcUpCmd := exec.Command("docker-compose", "-f", "-", "-p", appName, "up", "-d")
			dcUpCmd.Stdin = bytes.NewBuffer(data)
			dcUpCmd.Stdout = os.Stdout
			dcUpCmd.Stderr = os.Stderr
			dcUpCmd.Run()
		},
	}

	return cmd
}
