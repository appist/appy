package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/appist/appy/support"
)

// NewSecretCommand generates a cryptographically secure secret key which is typically used for cookie sessions.
func NewSecretCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "secret",
		Short: "Generate a cryptographically secure secret key which is typically used for cookie sessions.",
		Run: func(cmd *cobra.Command, args []string) {
			bytes := make([]byte, 64)
			if _, err := rand.Read(bytes); err != nil {
				support.Logger.Fatal(err)
			}

			fmt.Println(hex.EncodeToString(bytes))
		},
	}
}
