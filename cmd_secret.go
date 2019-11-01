package appy

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func newSecretCommand(logger *Logger) *Cmd {
	var length int

	cmd := &Cmd{
		Use:   "secret",
		Short: "Generates a cryptographically secure secret key for encrypting cookie, CSRF token and config",
		Run: func(cmd *Cmd, args []string) {
			bytes := make([]byte, length)

			if _, err := rand.Read(bytes); err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}

	cmd.Flags().IntVar(&length, "length", 64, "The byte length to generate, use 16 if you're generating for config encryption.")
	return cmd
}
