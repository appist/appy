package appy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// DBMigration contains database migration.
type DBMigration struct {
	File    string
	Version string
	Down    func(*DB) error
	DownTx  func(*DBTx) error
	Up      func(*DB) error
	UpTx    func(*DBTx) error
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

	t, err := template.New("migration").Parse(
		`package {{.DBName}}

import (
	"github.com/appist/appy"
	"{{.Module}}/pkg/app"
)

func init() {
	db := app.DB("{{.DBName}}")
	if db != nil {
		err := db.RegisterMigration{{if .Tx}}Tx{{end}}(
			// Up migration
			func(db *appy.DB{{if .Tx}}Tx{{end}}) error {
				_, err := db.Exec(` + "`" + "`" + `)
				return err
			},
			// Down migration
			func(db *appy.DB{{if .Tx}}Tx{{end}}) error {
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

	if err != nil {
		return []byte(""), err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data{
		DBName: dbname,
		Module: moduleName(),
		Tx:     tx,
	})

	if err != nil {
		return []byte(""), err
	}

	return tpl.Bytes(), err
}

func moduleName() string {
	wd, _ := os.Getwd()
	data, _ := ioutil.ReadFile(wd + "/go.mod")
	return parseModulePath(data)
}

func parseModulePath(mod []byte) string {
	var (
		slashSlash = []byte("//")
		moduleStr  = []byte("module")
	)

	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}

	return ""
}

func schemaDumpTpl(database, schema string) ([]byte, error) {
	type data struct {
		Database, Module, Schema string
	}

	t, err := template.New("schemaDump").Parse(
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

	if err != nil {
		return []byte(""), err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data{
		Database: database,
		Module:   moduleName(),
		Schema:   "\n" + schema,
	})

	if err != nil {
		return []byte(""), err
	}

	return tpl.Bytes(), err
}
