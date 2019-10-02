package cmd

import (
	"os"
	"os/exec"

	ah "github.com/appist/appy/http"

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
				panic(err)
			}

			os.RemoveAll(s.Config.HTTPSSLCertPath)
			cleanCmd := exec.Command("mkcert", "-uninstall")
			cleanCmd.Stdout = os.Stdout
			cleanCmd.Stderr = os.Stderr
			cleanCmd.Run()
		},
	}
}
