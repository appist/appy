package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewSSLSetupCommand generates the SSL certs and adds to the system trust store using `mkcert`, only available for
// debug build.
func NewSSLSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ssl:setup",
		Short: "Create and install the locally trusted SSL certs using `mkcert`, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			dir := filepath.Dir(config.HTTPSSLCertPath + "/ca.crt")
			_ = os.MkdirAll(dir, os.ModePerm)

			setupArgs := []string{"-install", "-cert-file", config.HTTPSSLCertPath + "/cert.pem", "-key-file", config.HTTPSSLCertPath + "/key.pem"}
			setupArgs = append(setupArgs, getIPHosts()...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
