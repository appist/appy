package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

// NewSecretCommand generates a cryptographically secure secret key for encrypting session cookie and CSRF token.
func NewSecretCommand() *AppCmd {
	return &AppCmd{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key for encrypting session cookie and CSRF token.",
		Run: func(cmd *AppCmd, args []string) {
			bytes := make([]byte, 64)

			if _, err := rand.Read(bytes); err != nil {
				log.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}
}
