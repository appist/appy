package cmd

import (
	"os"
	"os/exec"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
)

// NewSSLTeardownCommand uninstall the locally trusted SSL certs using `mkcert`.
func NewSSLTeardownCommand(logger *support.Logger, server *ah.Server) *Command {
	return &Command{
		Use:   "ssl:teardown",
		Short: "Uninstall the locally trusted SSL certs using `mkcert`",
		Run: func(cmd *Command, args []string) {
			if support.IsConfigErrored(server.Config(), logger) {
				os.Exit(-1)
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(server.Config().HTTPSSLCertPath)
			teardownCmd := exec.Command("mkcert", "-uninstall")
			teardownCmd.Stdout = os.Stdout
			teardownCmd.Stderr = os.Stderr
			teardownCmd.Run()
		},
	}
}
