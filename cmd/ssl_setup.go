package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/appist/appy/core"
)

// NewSSLSetupCommand generates and installs the locally trusted SSL certs using "mkcert", only available for debug build.
func NewSSLSetupCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "ssl:setup",
		Short: "Generates and installs the locally trusted SSL certs using \"mkcert\", only available for debug build.",
		Run: func(cmd *AppCmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				log.Fatal(err)
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
