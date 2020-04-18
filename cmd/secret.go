package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/appist/appy/support"
)

func newSecretCommand(logger *support.Logger) *Command {
	cmd := &Command{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key for encrypting cookie, CSRF token and config",
		Run: func(cmd *Command, args []string) {
			fmt.Println(hex.EncodeToString(support.GenerateRandomBytes(32)))
		},
	}

	return cmd
}
