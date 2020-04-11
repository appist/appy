package cmd

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/appist/appy/support"
	"github.com/joho/godotenv"
)

func newConfigDecCommand(config *support.Config, logger *support.Logger) *Command {
	return &Command{
		Use:   "config:dec <KEY>",
		Short: "Decrypt a config value using the secret in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`",
		Args:  ExactArgs(1),
		Run: func(cmd *Command, args []string) {
			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(support.ErrMissingMasterKey)
			}

			key := args[0]
			if !support.IsSnakeCase(strings.ToLower(key)) {
				logger.Fatal("invalid key format (e.g. HTTP_HOST)")
			}

			envMap, err := godotenv.Read(config.Path())
			if err != nil {
				logger.Fatal(err)
			}

			val, existed := envMap[key]
			if !existed {
				logger.Fatal(fmt.Sprintf("the key '%s' is not defined in '%s'", key, config.Path()))
			}

			ciphertext, err := hex.DecodeString(val)
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
