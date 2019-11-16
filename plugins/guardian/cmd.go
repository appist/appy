package guardian

import "github.com/appist/appy"

func newAppyGuardianInstallCommand() *appy.Command {
	return &appy.Command{
		Use:   "appy:guardian:install",
		Short: "Install everything needed for appy guardian plugin",
		Run: func(cmd *appy.Command, args []string) {
		},
	}
}
