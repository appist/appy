package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

// NewSecretConfigCommand generates a cryptographically secure secret key for encrypting application config.
func NewSecretConfigCommand() *AppCmd {
	return &AppCmd{
		Use:   "secret:config",
		Short: "Generate a cryptographically secure secret key for encrypting application config.",
		Run: func(cmd *AppCmd, args []string) {
			bytes := make([]byte, 16)

			if _, err := rand.Read(bytes); err != nil {
				log.Fatal(err)
			}

			secret := hex.EncodeToString(bytes)
			fmt.Println(secret[:len(secret)-1])
		},
	}
}
