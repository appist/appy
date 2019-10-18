package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

// NewSecretCommand generates a cryptographically secure secret key for encrypting cookie, CSRF token and config.
func NewSecretCommand() *AppCmd {
	var length int

	cmd := &AppCmd{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key for encrypting cookie, CSRF token and config.",
		Run: func(cmd *AppCmd, args []string) {
			bytes := make([]byte, length)

			if _, err := rand.Read(bytes); err != nil {
				log.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}

	cmd.Flags().IntVar(&length, "length", 64, "The byte length to generate, use 16 if you're generating for config encryption.")
	return cmd
}
