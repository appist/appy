package appy

import (
	"os"
	"os/exec"
)

func newSSLTeardownCommand(logger *Logger, server *Server) *Cmd {
	return &Cmd{
		Use:   "ssl:teardown",
		Short: `Uninstall the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(server.config, server.logger) {
				os.Exit(-1)
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(server.config.HTTPSSLCertPath)
			teardownCmd := exec.Command("mkcert", "-uninstall")
			teardownCmd.Stdout = os.Stdout
			teardownCmd.Stderr = os.Stderr
			teardownCmd.Run()
		},
	}
}
