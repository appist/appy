package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/appist/appy/support"
)

func newSecretCommand(logger *support.Logger) *Command {
	cmd := &Command{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key for encrypting cookie, CSRF token and config",
		Run: func(cmd *Command, args []string) {
			bytes := make([]byte, 32)

			if _, err := rand.Read(bytes); err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}

	return cmd
}
