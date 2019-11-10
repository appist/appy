package guardian

import "github.com/appist/appy"

func newAppyGuardianInstallCommand() *appy.Cmd {
	return &appy.Cmd{
		Use:   "appy:guardian:install",
		Short: "Install everything needed for appy guardian plugin",
		Run: func(cmd *appy.Cmd, args []string) {
		},
	}
}
