package appy

import (
	"os"
	"os/exec"
)

func newSSLSetupCommand(logger *Logger, server *Server) *Cmd {
	return &Cmd{
		Use:   "ssl:setup",
		Short: `Generate and install the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(server.config, server.logger) {
				os.Exit(-1)
			}

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
