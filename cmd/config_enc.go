package cmd

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/appist/appy/core"
	"github.com/appist/appy/support"
)

// NewConfigEncryptCommand encrypts a value using the AES algorithm, the master key that is used for the encryption can
// be passed in:
//
// 1. via the `APPY_ENV` environment variable which will be used to locate `config/<APPY_ENV>.key` file.
// 2. via the `APPY_MASTER_KEY` environment variable which will always take precedence over `config/<APPY_ENV>.key` file.
func NewConfigEncryptCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "config:enc",
		Short: "Encrypt a value using the AES algorithm.",
		Args:  ExactArgs(1),
		Run: func(cmd *AppCmd, args []string) {
			key, err := core.MasterKey()
			if err != nil {
				log.Fatal(err)
			}

			plaintext := []byte(args[0])
			ciphertext, err := support.Encrypt(plaintext, key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(ciphertext))
		},
	}
}
