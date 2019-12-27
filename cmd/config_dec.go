package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/appist/appy/support"
)

// NewConfigDecryptCommand decrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`.
func NewConfigDecryptCommand(config *support.Config, logger *support.Logger) *Command {
	return &Command{
		Use:   "config:dec",
		Short: "Decrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`",
		Args:  ExactArgs(1),
		Run: func(cmd *Command, args []string) {
			if support.IsConfigErrored(config, logger) {
				os.Exit(-1)
			}

			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(config.Errors())
			}

			ciphertext, err := hex.DecodeString(args[0])
			if err != nil {
				logger.Fatal(err)
			}

			decrypted, err := support.AESDecrypt(ciphertext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println(string(decrypted))
		},
	}
}
