package cmd

import (
	"os"
	"os/exec"

	"github.com/appist/appy/core"
)

// NewSSLCleanCommand uninstalls all the local SSL certs using "mkcert", only available for debug build.
func NewSSLCleanCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "ssl:clean",
		Short: "Uninstalls the locally trusted SSL certs using \"mkcert\", only available for debug build.",
		Run: func(cmd *AppCmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(s.Config.HTTPSSLCertPath)
			cleanCmd := exec.Command("mkcert", "-uninstall")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
		},
	}
}
