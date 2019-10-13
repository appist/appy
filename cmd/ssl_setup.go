package cmd

import (
	"os"
	"os/exec"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"

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
				support.Logger.Fatal(err)
			}

			os.MkdirAll(s.Config.HTTPSSLCertPath, os.ModePerm)
			setupArgs := []string{"-install", "-cert-file", s.Config.HTTPSSLCertPath + "/cert.pem", "-key-file", s.Config.HTTPSSLCertPath + "/key.pem"}
			hosts, _ := s.Hosts()
			setupArgs = append(setupArgs, hosts...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
