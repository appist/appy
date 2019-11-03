package appy

import (
	"os"
	"os/exec"
)

func newSSLCleanCommand(logger *Logger, server *Server) *Cmd {
	return &Cmd{
		Use:   "ssl:clean",
		Short: `Uninstall the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(server.config.HTTPSSLCertPath)
			cleanCmd := exec.Command("mkcert", "-uninstall")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
		},
	}
}
