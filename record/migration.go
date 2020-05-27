package record

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/appist/appy/support"
)

var (
	migratePath = "db/migrate/"
)

// Migration contains database migration.
type Migration struct {
	File    string
	Version string
	Down    func(DBer) error
	DownTx  func(Txer) error
	Up      func(DBer) error
	UpTx    func(Txer) error
}

func migrationFile() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	for {
		f, ok := frames.Next()
		if !ok {
			break
		}

		if !strings.Contains(f.Function, "/appist/appy") {
			return f.File
		}
	}

	return ""
}

func migrationVersion(name string) (string, error) {
	base := filepath.Base(name)
	splits := strings.Split(base, "_")
	_, err := time.Parse("20060102150405", splits[0])
	if len(splits) < 3 || err != nil {
		return "", fmt.Errorf("invalid filename '%q', a valid example: 20060102150405_create_users.go", base)
	}

	return splits[0], nil
}

func migrationTpl(dbname string, tx bool) ([]byte, error) {
	type data struct {
		DBName, Module string
		Tx             bool
	}

	t, _ := template.New("migration").Parse(
		`package {{.DBName}}

import (
	"github.com/appist/appy/record"
	"{{.Module}}/pkg/app"
)

func init() {
	db := app.DB("{{.DBName}}")

	if db != nil {
		err := db.RegisterMigration{{if .Tx}}Tx{{end}}(
			// Up migration
			func(db record.{{if .Tx}}Txer{{else}}DBer{{end}}) error {
				_, err := db.Exec(` + "`" + "`" + `)
				return err
			},
			// Down migration
			func(db record.{{if .Tx}}Txer{{else}}DBer{{end}}) error {
				_, err := db.Exec(` + "`" + "`" + `)
				return err
			},
		)

		if err != nil {
			app.Logger.Fatal(err)
		}
	}
}
`)

	var tpl bytes.Buffer
	tpl.Write([]byte(""))

	err := t.Execute(&tpl, data{
		DBName: dbname,
		Module: support.ModuleName(),
		Tx:     tx,
	})

	return tpl.Bytes(), err
}

func schemaDumpTpl(database, schema string) ([]byte, error) {
	type data struct {
		Database, Module, Schema string
	}

	t, _ := template.New("schemaDump").Parse(
		`package {{.Database}}

import (
	"{{.Module}}/pkg/app"
)

func init() {
	db := app.DB("{{.Database}}")

	if db != nil {
		db.SetSchema(` + "`" +
			"{{.Schema}}" +
			"`)" +
			`
	}
}
`)

	var tpl bytes.Buffer
	tpl.Write([]byte(""))

	err := t.Execute(&tpl, data{
		Database: database,
		Module:   support.ModuleName(),
		Schema:   "\n" + schema,
	})

	return tpl.Bytes(), err
}
