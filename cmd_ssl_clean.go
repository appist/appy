package appy

import (
	"os"
	"os/exec"
)

func newSSLCleanCommand(config *Config, logger *Logger) *Cmd {
	return &Cmd{
		Use:   "ssl:clean",
		Short: `Uninstalls the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(config.HTTPSSLCertPath)
			cleanCmd := exec.Command("mkcert", "-uninstall")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
		},
	}
}
