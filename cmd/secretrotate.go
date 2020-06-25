package cmd

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/appist/appy/support"
	"github.com/joho/godotenv"
)

func newSecretRotateCommand(asset *support.Asset, config *support.Config, logger *support.Logger) *Command {
	return &Command{
		Use:   "secret:rotate <OLD_SECRET> <NEW_SECRET>",
		Short: "Rotate the secret that is used to encrypt/decrypt the configs (only available in debug build)",
		Args:  ExactArgs(2),
		Run: func(cmd *Command, args []string) {
			envMap, err := godotenv.Read(config.Path())
			if err != nil {
				logger.Fatal(err)
			}

			envKeys := make([]string, 0, len(envMap))
			for key := range envMap {
				envKeys = append(envKeys, key)
			}
			sort.Strings(envKeys)

			newData := ""
			prevKey := ""
			for _, key := range envKeys {
				ciphertext, err := hex.DecodeString(envMap[key])
				if err != nil {
					logger.Fatal(err)
				}

				decrypted, err := support.AESDecrypt(ciphertext, []byte(args[0]))
				if err != nil {
					logger.Fatal(err)
				}

				encrypted, err := support.AESEncrypt(decrypted, []byte(args[1]))
				if err != nil {
					logger.Fatal(err)
				}

				if prevKey != "" && prevKey[0] != key[0] {
					newData += "\n"
				}

				prevKey = key
				newData += fmt.Sprintf("%s=%s\n", key, hex.EncodeToString(encrypted))
			}

			err = ioutil.WriteFile(config.Path(), []byte(newData), 0)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Printf("Successfully re-encrypted '%s' with the new secret key!\n", config.Path())

			keyFile := fmt.Sprintf("%s/%s.key", asset.Layout().Config(), config.AppyEnv)
			info, err := os.Stat(keyFile)
			if os.IsNotExist(err) || info.IsDir() {
				return
			}

			err = ioutil.WriteFile(keyFile, []byte(args[1]), 0)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}
}
