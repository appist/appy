//+build !test

package appy

import (
	"os"
	"os/exec"
)

func newSSLTeardownCommand(logger *Logger, server *Server) *Command {
	return &Command{
		Use:   "ssl:teardown",
		Short: "Uninstall the locally trusted SSL certs using `mkcert`",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(server.Config().HTTPSSLCertPath)
			teardownCmd := exec.Command("mkcert", "-uninstall")
			teardownCmd.Stdout = os.Stdout
			teardownCmd.Stderr = os.Stderr
			teardownCmd.Run()
		},
	}
}
