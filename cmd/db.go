package cmd

import (
	"fmt"
	"os"

	"github.com/appist/appy/core"
)

func checkProtectedEnvs(config core.AppConfig) {
	if config.AppyEnv == "production" {
		fmt.Printf("You are attempting to run a destructive action against your '%s' database.\n", config.AppyEnv)
		os.Exit(-1)
	}
}
