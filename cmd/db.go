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

func dbCreate(db *core.AppDb) ([]string, []string) {
	var errs, msgs []string
	dbName := db.Config.Database
	_, err := db.Handler.Exec("CREATE DATABASE \"?\"", dbName)
	if err != nil {
		errs = append(errs, err.Error())
	}
	msgs = append(msgs, fmt.Sprintf("Successfully created database \"%s\".", dbName))

	_, err = db.Handler.Exec("CREATE DATABASE \"?\"", dbName+"_test")
	if err != nil {
		errs = append(errs, err.Error())
	}

	msgs = append(msgs, fmt.Sprintf("Successfully created database \"%s\".", dbName+"_test"))
	return msgs, errs
}

func dbDrop(db *core.AppDb) ([]string, []string) {
	var errs, msgs []string
	dbName := db.Config.Database
	_, err := db.Handler.Exec("DROP DATABASE \"?\"", dbName)
	if err != nil {
		errs = append(errs, err.Error())
	}
	msgs = append(msgs, fmt.Sprintf("Successfully dropped database \"%s\".", dbName))

	_, err = db.Handler.Exec("DROP DATABASE \"?\"", dbName+"_test")
	if err != nil {
		errs = append(errs, err.Error())
	}
	msgs = append(msgs, fmt.Sprintf("Successfully dropped database \"%s\".", dbName+"_test"))

	return msgs, errs
}
