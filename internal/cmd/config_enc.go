package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	appysupport "github.com/appist/appy/internal/support"
)

// NewConfigEncryptCommand encrypt a value using the AES algorithm.
func NewConfigEncryptCommand(config *appysupport.Config, logger *appysupport.Logger) *Command {
	return &Command{
		Use:   "config:enc",
		Short: "Encrypt a value using the AES algorithm",
		Args:  ExactArgs(1),
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) {
				os.Exit(-1)
			}

			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(config.Errors())
			}

			plaintext := []byte(args[0])
			ciphertext, err := appysupport.AESEncrypt(plaintext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(ciphertext))
		},
	}
}
