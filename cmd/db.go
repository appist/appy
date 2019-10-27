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

func dbCreate(db *core.AppDb) []string {
	var errs []string
	dbName := db.Config.Database
	_, err := db.Handler.Exec(`CREATE DATABASE ?`, core.SafeQuery(dbName))
	if err != nil {
		errs = append(errs, err.Error())
	}

	_, err = db.Handler.Exec(`CREATE DATABASE ?`, core.SafeQuery(dbName+"_test"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	return errs
}

func dbDrop(db *core.AppDb) []string {
	var errs []string
	dbName := db.Config.Database
	_, err := db.Handler.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '?'`, core.SafeQuery(dbName))
	if err != nil {
		errs = append(errs, err.Error())
	}

	_, err = db.Handler.Exec(`DROP DATABASE ?`, core.SafeQuery(dbName))
	if err != nil {
		errs = append(errs, err.Error())
	}

	_, err = db.Handler.Exec(`DROP DATABASE ?`, core.SafeQuery(dbName+"_test"))
	if err != nil {
		errs = append(errs, err.Error())
	}

	return errs
}
