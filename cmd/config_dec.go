package cmd

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/appist/appy/core"
	"github.com/appist/appy/support"
)

// NewConfigDecryptCommand decrypts a value using the AES algorithm, the master key that is used for the decryption can
// be passed in:
//
// 1. via the `APPY_ENV` environment variable which will be used to locate `config/<APPY_ENV>.key` file.
// 2. via the `APPY_MASTER_KEY` environment variable which will always take precedence over `config/<APPY_ENV>.key` file.
func NewConfigDecryptCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "config:dec",
		Short: "Decrypt a value using the AES algorithm.",
		Args:  ExactArgs(1),
		Run: func(cmd *AppCmd, args []string) {
			key, err := core.MasterKey()
			if err != nil {
				log.Fatal(err)
			}

			ciphertext, err := hex.DecodeString(args[0])
			decrypted, err := support.Decrypt(ciphertext, key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(decrypted))
		},
	}
}
