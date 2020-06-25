package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/appist/appy/support"
	"github.com/joho/godotenv"
)

func newConfigEncCommand(config *support.Config, logger *support.Logger) *Command {
	return &Command{
		Use:   "config:enc <KEY> <VALUE>",
		Short: "Encrypt a config value using the secret in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY` (only available in debug build)",
		Args:  ExactArgs(2),
		Run: func(cmd *Command, args []string) {
			masterKey := config.MasterKey()
			if masterKey == nil {
				logger.Fatal(support.ErrMissingMasterKey)
			}

			key := args[0]
			value := args[1]
			if !support.IsSnakeCase(strings.ToLower(key)) {
				logger.Fatal("please provide key in upper snake_case, e.g. HTTP_HOST")
			}

			envMap, err := godotenv.Read(config.Path())
			if err != nil {
				logger.Fatal(err)
			}

			_, existed := envMap[key]
			if existed {
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("'%s' key already existed in '%s', do you want to overwrite? (y/N): ", key, config.Path())
				answer, _ := reader.ReadString('\n')
				answer = strings.Trim(strings.ToLower(answer), "\n")

				if answer != "y" {
					os.Exit(0)
				}
			}

			plaintext := []byte(value)
			ciphertext, err := support.AESEncrypt(plaintext, masterKey)
			if err != nil {
				logger.Fatal(err)
			}
			envMap[key] = hex.EncodeToString(ciphertext)

			envKeys := make([]string, 0, len(envMap))
			for k := range envMap {
				envKeys = append(envKeys, k)
			}
			sort.Strings(envKeys)

			newData := ""
			prevKey := ""
			for _, k := range envKeys {
				if prevKey != "" && prevKey[0] != k[0] {
					newData += "\n"
				}

				prevKey = k
				newData += fmt.Sprintf("%s=%s\n", k, envMap[k])
			}

			err = ioutil.WriteFile(config.Path(), []byte(newData), 0)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Printf("Successfully stored the enrypted value for '%s' key into '%s'!\n", key, config.Path())
		},
	}
}
