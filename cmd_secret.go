//+build !test

package appy

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func newSecretCommand(logger *Logger) *Command {
	var length int

	cmd := &Command{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key for encrypting cookie, CSRF token and config",
		Run: func(cmd *Command, args []string) {
			bytes := make([]byte, length)

			if _, err := rand.Read(bytes); err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}

	cmd.Flags().IntVar(&length, "length", 64, "The byte length to generate, use 16 if you're generating for config encryption")
	return cmd
}
