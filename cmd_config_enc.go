package appy

import (
	"encoding/hex"
	"fmt"
)

func newConfigEncryptCommand(config *Config, logger *Logger, support *Support) *Cmd {
	return &Cmd{
		Use:   "config:enc",
		Short: "Encrypts a value using the AES algorithm",
		Args:  ExactArgs(1),
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)

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
