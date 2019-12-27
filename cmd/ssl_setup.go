package cmd

import (
	"os"
	"os/exec"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
)

// NewSSLSetupCommand generate and install the locally trusted SSL certs using `mkcert`.
func NewSSLSetupCommand(logger *support.Logger, server *ah.Server) *Command {
	return &Command{
		Use:   "ssl:setup",
		Short: "Generate and install the locally trusted SSL certs using `mkcert`",
		Run: func(cmd *Command, args []string) {
			if support.IsConfigErrored(server.Config(), logger) {
				os.Exit(-1)
			}

			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.MkdirAll(server.Config().HTTPSSLCertPath, os.ModePerm)
			setupArgs := []string{"-install", "-cert-file", server.Config().HTTPSSLCertPath + "/cert.pem", "-key-file", server.Config().HTTPSSLCertPath + "/key.pem"}
			hosts, _ := server.Hosts()
			setupArgs = append(setupArgs, hosts...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}