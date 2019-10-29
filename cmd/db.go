package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/appist/appy/core"
)

func checkProtectedEnvs(config core.AppConfig) {
	if config.AppyEnv == "production" {
		fmt.Printf("You are attempting to run a destructive action against your '%s' database.\n", config.AppyEnv)
		os.Exit(-1)
	}
}

func dumpSchema(name string, db *core.AppDb) error {
	path := "db/migrations/" + name
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	// A trick to utilise url package to parse our database address in a form like 0.0.0.0:5432
	u, err := url.Parse("https://" + db.Config.Addr)
	if err != nil {
		return err
	}

	var (
		outBytes bytes.Buffer
		out      string
	)
	dumpArgs := []string{
		"-s", "-x", "-O", "--no-comments",
		"-d", db.Config.Database,
		"-n", db.Config.Schema,
		"-h", u.Hostname(),
		"-p", u.Port(),
		"-U", db.Config.User,
	}
	dumpCmd := exec.Command("pg_dump", dumpArgs...)
	dumpCmd.Env = os.Environ()
	dumpCmd.Env = append(dumpCmd.Env, []string{"PGPASSWORD=" + db.Config.Password}...)
	dumpCmd.Stdout = &outBytes
	dumpCmd.Stderr = os.Stderr
	dumpCmd.Run()
	out = outBytes.String()
	out = regexp.MustCompile(`(?i)--\n-- postgresql database dump.*\n--\n\n`).ReplaceAllString(out, "")
	out = regexp.MustCompile(`(?i)--\ dumped.*\n(\n)?`).ReplaceAllString(out, "")
	out = strings.Trim(out, "\n")
	out += fmt.Sprintf("\n\nINSERT INTO %s.%s (version) VALUES\n", db.Config.Schema, db.Config.SchemaMigrationsTable)

	var schemaMigrations []core.AppDbSchemaMigration
	_, err = db.Handler.Query(&schemaMigrations, `SELECT version FROM ?.? ORDER BY version ASC`, core.SafeQuery(db.Config.Schema), core.SafeQuery(db.Config.SchemaMigrationsTable))
	if err != nil {
		return err
	}

	for idx, m := range schemaMigrations {
		out += "('" + m.Version + "')"

		if idx == len(schemaMigrations)-1 {
			out += ";\n"
		} else {
			out += ",\n"
		}
	}

	fn := "db/migrations/" + name + "/schema.go"
	tpl, _ := schemaDumpTpl(name, out)
	err = ioutil.WriteFile(fn, tpl, 0644)

	return err
}

func schemaDumpTpl(database, schema string) ([]byte, error) {
	type data struct {
		Database string
		Schema   string
	}

	t, err := template.New("schemaDump").Parse(
		`package {{.Database}}

import (
	"github.com/appist/appy"
)

func init() {
	appy.Db["{{.Database}}"].Schema = ` + "`" +
			"{{.Schema}}" +
			"`" +
			`
}
`)

	if err != nil {
		return []byte(""), err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data{
		Database: database,
		Schema:   "\n" + schema,
	})

	if err != nil {
		return []byte(""), err
	}

	return tpl.Bytes(), err
}
