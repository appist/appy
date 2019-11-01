package appy

import (
	"encoding/hex"
	"fmt"
)

func newConfigDecryptCommand(config *Config, logger *Logger, support *Support) *Cmd {
	return &Cmd{
		Use:   "config:dec",
		Short: "Decrypts a value using the AES algorithm",
		Args:  ExactArgs(1),
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(config, logger)

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
