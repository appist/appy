package cmd

import (
	"os"
	"os/exec"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
)

func newSSLTearDownCommand(logger *support.Logger, server *pack.Server) *Command {
	return &Command{
		Use:   "ssl:teardown",
		Short: "Uninstall the locally trusted SSL certs using `mkcert`",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
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
