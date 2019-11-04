package appy

import (
	"encoding/hex"
	"fmt"
	"os"
)

func newConfigEncryptCommand(config *Config, logger *Logger) *Cmd {
	return &Cmd{
		Use:   "config:enc",
		Short: "Encrypt a value using the AES algorithm",
		Args:  ExactArgs(1),
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(config, logger) {
				os.Exit(-1)
			}

			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(config.Errors())
			}

			plaintext := []byte(args[0])
			ciphertext, err := AESEncrypt(plaintext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(ciphertext))
		},
	}
}
