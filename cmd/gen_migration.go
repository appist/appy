package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/appist/appy/core"
	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

// NewGenMigrationCommand generates a database migration file for the current environment.
func NewGenMigrationCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	var (
		tx     bool
		target string
	)

	cmd := &AppCmd{
		Use:   "gen:migration NAME",
		Short: "Generates a database migration file for the current environment.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || !support.IsPascalCase(args[0]) {
				return errors.New("please provide migration name in pascal case, e.g. CreateUsers")
			}

			return nil
		},
		Run: func(cmd *AppCmd, args []string) {
			if _, ok := dbMap[target]; !ok {
				logger.Fatal(fmt.Errorf("database '%s' is not defined for the current environment", target))
			}

			path := "db/migrations/" + target
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				logger.Fatal(err)
			}

			ts := time.Now()
			fn := path + "/" + ts.Format("20060102150405") + "_" + support.ToSnakeCase(args[0]) + ".go"
			tpl, err := migrationTpl(target, tx)
			if err != nil {
				logger.Fatal(err)
			}

			err = ioutil.WriteFile(fn, tpl, 0644)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infof("Created migration for '%s' database in '%s'.", target, fn)
		},
	}

	cmd.Flags().StringVar(&target, "database", "primary", "The target database to generate the migration file for.")
	cmd.Flags().BoolVar(&tx, "tx", true, "Indicate whether or not the migration should execute in a transaction.")
	return cmd
}

func migrationTpl(database string, tx bool) ([]byte, error) {
	type data struct {
		Database string
		Tx       bool
	}

	t, err := template.New("migration").Parse(
		`package {{.Database}}

import (
	"github.com/appist/appy"
)

func init() {
	appy.Db["{{.Database}}"].RegisterMigration{{if .Tx}}Tx{{end}}(
		// Up migration
		func(db *appy.AppDb) error {
			return nil
		},

		// Down migration
		func(db *appy.AppDb) error {
			return nil
		},
	)
}
`)

	if err != nil {
		return []byte(""), err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data{
		Database: database,
		Tx:       tx,
	})

	if err != nil {
		return []byte(""), err
	}

	return tpl.Bytes(), err
}
