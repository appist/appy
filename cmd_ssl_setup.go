//+build !test

package appy

import (
	"os"
	"os/exec"
)

func newSSLSetupCommand(logger *Logger, server *Server) *Command {
	return &Command{
		Use:   "ssl:setup",
		Short: "Generate and install the locally trusted SSL certs using `mkcert`",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.MkdirAll(server.Config().HTTPSSLCertPath, os.ModePerm)
			setupArgs := []string{"-install", "-cert-file", server.Config().HTTPSSLCertPath + "/cert.pem", "-key-file", server.Config().HTTPSSLCertPath + "/key.pem"}
			hosts, _ := server.Hosts()
			if !server.support.ArrayContains(hosts, "0.0.0.0") {
				hosts = append(hosts, "0.0.0.0")
			}

			setupArgs = append(setupArgs, hosts...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
