package appy

import (
	"os"
	"os/exec"
)

func newSSLSetupCommand(logger *Logger, server *Server) *Cmd {
	return &Cmd{
		Use:   "ssl:setup",
		Short: `Generates and installs the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.MkdirAll(server.config.HTTPSSLCertPath, os.ModePerm)
			setupArgs := []string{"-install", "-cert-file", server.config.HTTPSSLCertPath + "/cert.pem", "-key-file", server.config.HTTPSSLCertPath + "/key.pem"}
			hosts, _ := server.Hosts()
			setupArgs = append(setupArgs, hosts...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
