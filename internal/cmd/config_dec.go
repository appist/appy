package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	appysupport "github.com/appist/appy/internal/support"
)

// NewConfigDecryptCommand decrypt a value using the AES algorithm.
func NewConfigDecryptCommand(config *appysupport.Config, logger *appysupport.Logger) *Command {
	return &Command{
		Use:   "config:dec",
		Short: "Decrypt a value using the AES algorithm",
		Args:  ExactArgs(1),
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) {
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

			decrypted, err := appysupport.AESDecrypt(ciphertext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println(string(decrypted))
		},
	}
}
