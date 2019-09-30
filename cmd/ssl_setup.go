package cmd

import (
	ah "appist/appy/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewSSLSetupCommand generates the SSL certs and adds to the system trust store using `mkcert`, only available for
// debug build.
func NewSSLSetupCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "ssl:setup",
		Short: "Create and install the locally trusted SSL certs using `mkcert`, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			dir := filepath.Dir(s.Config.HTTPSSLCertPath + "/ca.crt")
			_ = os.MkdirAll(dir, os.ModePerm)

			setupArgs := []string{"-install", "-cert-file", s.Config.HTTPSSLCertPath + "/cert.pem", "-key-file", s.Config.HTTPSSLCertPath + "/key.pem"}
			setupArgs = append(setupArgs, getIPHosts(s)...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
