package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/appist/appy/support"
)

// NewConfigEncryptCommand encrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`.
func NewConfigEncryptCommand(config *support.Config, logger *support.Logger) *Command {
	return &Command{
		Use:   "config:enc",
		Short: "Encrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`",
		Args:  ExactArgs(1),
		Run: func(cmd *Command, args []string) {
			if support.IsConfigErrored(config, logger) {
				os.Exit(-1)
			}

			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(config.Errors())
			}

			plaintext := []byte(args[0])
			ciphertext, err := support.AESEncrypt(plaintext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(ciphertext))
		},
	}
}
