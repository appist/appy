package appy

import (
	"os"
	"os/exec"
)

func newSSLSetupCommand(config *Config, logger *Logger, s *Server) *Cmd {
	return &Cmd{
		Use:   "ssl:setup",
		Short: `Generates and installs the locally trusted SSL certs using "mkcert"`,
		Run: func(cmd *Cmd, args []string) {
			_, err := exec.LookPath("mkcert")
			if err != nil {
				logger.Fatal(err)
			}

			os.MkdirAll(config.HTTPSSLCertPath, os.ModePerm)
			setupArgs := []string{"-install", "-cert-file", config.HTTPSSLCertPath + "/cert.pem", "-key-file", config.HTTPSSLCertPath + "/key.pem"}
			hosts, _ := s.Hosts(config)
			setupArgs = append(setupArgs, hosts...)
			setupCmd := exec.Command("mkcert", setupArgs...)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			setupCmd.Run()
		},
	}
}
