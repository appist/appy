package cmd

import (
	"os"
	"os/exec"

	appyhttp "github.com/appist/appy/internal/http"
	appysupport "github.com/appist/appy/internal/support"
)

// NewSSLTeardownCommand Uninstall the locally trusted SSL certs using "mkcert".
func NewSSLTeardownCommand(logger *appysupport.Logger, s *appyhttp.Server) *Command {
	return &Command{
		Use:   "ssl:teardown",
		Short: `Uninstall the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(s.Config(), logger) {
				os.Exit(-1)
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.RemoveAll(s.Config().HTTPSSLCertPath)
			teardownCmd := exec.Command("mkcert", "-uninstall")
			teardownCmd.Stdout = os.Stdout
			teardownCmd.Stderr = os.Stderr
			teardownCmd.Run()
		},
	}
}
