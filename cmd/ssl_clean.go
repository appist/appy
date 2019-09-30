package cmd

import (
	ah "appist/appy/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewSSLCleanCommand removes/uninstalls all the local SSL certs using `mkcert`, only available for debug build.
func NewSSLCleanCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "ssl:clean",
		Short: "Uninstall and clean up the locally trusted SSL certs using `mkcert`, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			dir := filepath.Dir(s.Config.HTTPSSLCertPath + "/ca.crt")
			_ = os.RemoveAll(dir)

			cleanCmd := exec.Command("mkcert", "-uninstall")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
		},
	}
}
