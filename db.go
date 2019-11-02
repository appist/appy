package appy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-pg/pg/v9"
)

type (
	// DbConn represents a single database connection rather than a pool of database
	// connections. Prefer running queries from DB unless there is a specific
	// need for a continuous single database connection.
	//
	// A Conn must call Close to return the connection to the database pool
	// and may do so concurrently with a running query.
	//
	// After a call to Close, all operations on the connection fail.
	DbConn = pg.Conn

	// DbHandle is a database handle representing a pool of zero or more underlying connections. It's safe
	// for concurrent use by multiple goroutines.
	DbHandle = pg.DB

	// DbHandleTx is an in-progress database transaction. It is safe for concurrent use by multiple goroutines.
	DbHandleTx = pg.Tx

	// DbQueryEvent keeps the query event information.
	DbQueryEvent = pg.QueryEvent

	// DbConfig contains the database config.
	DbConfig struct {
		pg.Options
		Replica               bool
		SchemaSearchPath      string
		SchemaMigrationsTable string
	}

	Dber interface {
	}

	// Db provides the functionality to communicate with the database.
	Db struct {
		config     DbConfig
		handle     *DbHandle
		logger     *Logger
		migrations []*DbMigration
		mu         sync.Mutex
		schema     string
	}

	// DbMigration contains database migration.
	DbMigration struct {
		File    string
		Version string
		Down    func(*DbHandle) error
		DownTx  func(*DbHandleTx) error
		Up      func(*DbHandle) error
		UpTx    func(*DbHandleTx) error
	}

	// SchemaMigration
	SchemaMigration struct {
		Version string
	}
)

var (
	// DbArray accepts a slice and returns a wrapper for working with PostgreSQL array data type.
	// For struct fields you can use array tag:
	//
	//    Emails  []string `pg:",array"`
	DbArray = pg.Array

	// DbHstore accepts a map and returns a wrapper for working with hstore data type.
	// For struct fields you can use hstore tag:
	//
	//    Attrs map[string]string `pg:",hstore"`
	DbHstore = pg.Hstore

	// DbSafeQuery replaces any placeholders found in the query.
	DbSafeQuery = pg.SafeQuery

	// DbScan returns ColumnScanner that copies the columns in the row into the values.
	DbScan = pg.Scan
)

// Connect connects to a database using provided options and assign the database Handler which is safe for concurrent
// use by multiple goroutines and maintains its own connection pool.
func (db *Db) Connect(sameDb bool) error {
	opts := db.config.Options
	if !sameDb {
		opts.Database = "postgres"
	}

	db.handle = pg.Connect(&opts)
	db.handle.AddQueryHook(db.logger)
	_, err := db.handle.Exec("SELECT 1 /* appy framework */")

	return err
}

// Close closes the database connection and release any open resources. It is rare to Close a DB, as the DB handle is
// meant to be long-lived and shared between many goroutines.
func (db *Db) Close() error {
	return nil
}

// Create creates the database.
func (db *Db) Create() []error {
	var errs []error
	_, err := db.handle.Exec(`CREATE DATABASE ?`, DbSafeQuery(db.config.Database))
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.handle.Exec(`CREATE DATABASE ?`, DbSafeQuery(db.config.Database+"_test"))
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Drop creates the database.
func (db *Db) Drop() []error {
	var errs []error
	_, err := db.handle.Exec(
		`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '?'`,
		DbSafeQuery(db.config.Database),
	)
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.handle.Exec(`DROP DATABASE ?`, DbSafeQuery(db.config.Database))
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.handle.Exec(`DROP DATABASE ?`, DbSafeQuery(db.config.Database+"_test"))
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Migrate runs migrations for the current environment that have not run yet.
func (db *Db) Migrate() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	tx, err := db.handle.Begin()
	defer tx.Close()
	if err != nil {
		return err
	}

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return err
	}

	for _, m := range db.migrations {
		if !ArrayContains(migratedVersions, m.Version) {
			if m.UpTx != nil {
				err = m.UpTx(tx)
				if err != nil {
					return err
				}

				err = db.addSchemaMigration(nil, tx, m)
				if err != nil {
					return err
				}

				continue
			}

			err = m.Up(db.handle)
			if err != nil {
				return err
			}

			err = db.addSchemaMigration(db.handle, nil, m)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// MigrateStatus returns the migration status for the current environment.
func (db *Db) MigrateStatus() ([][]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return nil, err
	}

	tx, err := db.handle.Begin()
	defer tx.Close()
	if err != nil {
		return nil, err
	}

	var migrationStatus [][]string
	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return nil, err
	}

	wd, _ := os.Getwd()
	for _, m := range db.migrations {
		status := "down"
		if ArrayContains(migratedVersions, m.Version) {
			status = "up"
		}

		migrationStatus = append(migrationStatus, []string{status, m.Version, strings.ReplaceAll(m.File, wd+"/", "")})
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return migrationStatus, nil
}

// Rollback rolls back the last migration for the current environment.
func (db *Db) Rollback() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	tx, err := db.handle.Begin()
	defer tx.Close()
	if err != nil {
		return err
	}

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return err
	}

	if len(migratedVersions) > 0 {
		for i := len(db.migrations) - 1; i > -1; i-- {
			m := db.migrations[i]

			if migratedVersions[len(migratedVersions)-1] == m.Version {
				if m.DownTx != nil {
					err = m.DownTx(tx)
					if err != nil {
						return err
					}

					err = db.removeSchemaMigration(nil, tx, m)
					if err != nil {
						return err
					}

					continue
				}

				err = m.Down(db.handle)
				if err != nil {
					return err
				}

				err = db.removeSchemaMigration(db.handle, nil, m)
				if err != nil {
					return err
				}

				break
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// RegisterMigration registers the up/down migrations that won't be executed in transaction.
func (db *Db) RegisterMigration(up func(*DbHandle) error, down func(*DbHandle) error) {
	err := db.registerMigration(up, down, nil, nil)
	if err != nil {
		db.logger.Fatal(err)
	}
}

// RegisterMigrationTx registers the up/down migrations that will be executed in transaction.
func (db *Db) RegisterMigrationTx(upTx func(*DbHandleTx) error, downTx func(*DbHandleTx) error) {
	err := db.registerMigration(nil, nil, upTx, downTx)
	if err != nil {
		db.logger.Fatal(err)
	}
}

func (db *Db) addMigration(newMigration *DbMigration) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.migrations = append(db.migrations, newMigration)
}

func (db *Db) addSchemaMigration(handler *DbHandle, handlerTx *DbHandleTx, targetMigration *DbMigration) error {
	query := `INSERT INTO ?.? (version) VALUES (?)`
	if handler != nil {
		_, err := handler.Exec(
			query,
			DbSafeQuery(db.config.SchemaSearchPath),
			DbSafeQuery(db.config.SchemaMigrationsTable),
			DbSafeQuery(targetMigration.Version),
		)
		return err
	}

	_, err := handlerTx.Exec(
		query,
		DbSafeQuery(db.config.SchemaSearchPath),
		DbSafeQuery(db.config.SchemaMigrationsTable),
		DbSafeQuery(targetMigration.Version),
	)
	return err
}

func (db *Db) removeSchemaMigration(handler *DbHandle, handlerTx *DbHandleTx, targetMigration *DbMigration) error {
	query := `DELETE FROM ?.? WHERE version = '?'`
	if handler != nil {
		_, err := handler.Exec(
			query,
			DbSafeQuery(db.config.SchemaSearchPath),
			DbSafeQuery(db.config.SchemaMigrationsTable),
			DbSafeQuery(targetMigration.Version),
		)
		return err
	}

	_, err := handlerTx.Exec(
		query,
		DbSafeQuery(db.config.SchemaSearchPath),
		DbSafeQuery(db.config.SchemaMigrationsTable),
		DbSafeQuery(targetMigration.Version),
	)
	return err
}

func (db *Db) dumpSchema(name string) error {
	var (
		outBytes bytes.Buffer
		out      string
	)

	path := "db/migrations/" + name
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	// A trick to utilise url package to parse our database address in a form like 0.0.0.0:5432
	u, err := url.Parse("https://" + db.config.Addr)
	if err != nil {
		return err
	}

	err = db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	_, err = exec.LookPath("pg_dump")
	if err != nil {
		return err
	}

	dumpArgs := []string{
		"-s", "-x", "-O", "--no-comments",
		"-d", db.config.Database,
		"-n", db.config.SchemaSearchPath,
		"-h", u.Hostname(),
		"-p", u.Port(),
		"-U", db.config.User,
	}
	dumpCmd := exec.Command("pg_dump", dumpArgs...)
	dumpCmd.Env = os.Environ()
	dumpCmd.Env = append(dumpCmd.Env, []string{"PGPASSWORD=" + db.config.Password}...)
	dumpCmd.Stdout = &outBytes
	dumpCmd.Stderr = os.Stderr
	dumpCmd.Run()
	out = outBytes.String()
	out = regexp.MustCompile(`(?i)--\n-- postgresql database dump.*\n--\n\n`).ReplaceAllString(out, "")
	out = regexp.MustCompile(`(?i)--\ dumped.*\n(\n)?`).ReplaceAllString(out, "")
	out = regexp.MustCompile(`(?i)create\ extension`).ReplaceAllString(out, "CREATE EXTENSION IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ schema`).ReplaceAllString(out, "CREATE SCHEMA IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ sequence`).ReplaceAllString(out, "CREATE SEQUENCE IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ table`).ReplaceAllString(out, "CREATE TABLE IF NOT EXISTS")
	out = strings.Trim(out, "\n")

	var schemaMigrations []SchemaMigration
	_, err = db.handle.Query(
		&schemaMigrations,
		`SELECT version FROM ?.? ORDER BY version ASC`,
		DbSafeQuery(db.config.SchemaSearchPath),
		DbSafeQuery(db.config.SchemaMigrationsTable),
	)

	if err != nil {
		return err
	}

	if len(schemaMigrations) > 0 {
		out += fmt.Sprintf("\n\nINSERT INTO %s.%s (version) VALUES\n", db.config.SchemaSearchPath, db.config.SchemaMigrationsTable)

		for idx, m := range schemaMigrations {
			out += "('" + m.Version + "')"

			if idx == len(schemaMigrations)-1 {
				out += ";\n"
			} else {
				out += ",\n"
			}
		}
	}

	tpl, err := schemaDumpTpl(name, out)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path+"/schema.go", tpl, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (db *Db) ensureSchemaMigrationsTable() error {
	count, err := db.handle.
		Model().
		Table("pg_tables").
		Where("schemaname = '?'", DbSafeQuery(db.config.SchemaSearchPath)).
		Where("tablename = '?'", DbSafeQuery(db.config.SchemaMigrationsTable)).
		Count()

	if err != nil {
		return err
	}

	if count < 1 {
		_, err = db.handle.Exec(`CREATE SCHEMA IF NOT EXISTS ?`, DbSafeQuery(db.config.SchemaSearchPath))
		if err != nil {
			return err
		}

		_, err = db.handle.Exec(`CREATE TABLE ? (version VARCHAR PRIMARY KEY)`, DbSafeQuery(db.config.SchemaMigrationsTable))
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) migratedVersions(tx *DbHandleTx) ([]string, error) {
	var schemaMigrations []SchemaMigration
	_, err := tx.Query(
		&schemaMigrations,
		`SELECT version FROM ?.? ORDER BY version ASC`,
		DbSafeQuery(db.config.SchemaSearchPath),
		DbSafeQuery(db.config.SchemaMigrationsTable),
	)
	if err != nil {
		return nil, err
	}

	migratedVersions := []string{}
	for _, m := range schemaMigrations {
		migratedVersions = append(migratedVersions, m.Version)
	}

	return migratedVersions, nil
}

func (db *Db) registerMigration(up func(*DbHandle) error, down func(*DbHandle) error, upTx func(*DbHandleTx) error, downTx func(*DbHandleTx) error) error {
	file := migrationFile()
	version, err := migrationVersion(file)
	if err != nil {
		return err
	}

	db.addMigration(&DbMigration{
		File:    file,
		Version: version,
		Down:    down,
		DownTx:  downTx,
		Up:      up,
		UpTx:    upTx,
	})

	return nil
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
		err := fmt.Errorf("invalid filename '%q', a valid example: 20060102150405_create_users.go", base)
		return "", err
	}

	return splits[0], nil
}

// Schema returns the database schema.
func (db *Db) Schema() string {
	return db.schema
}

// SetSchema stores the database schema.
func (db *Db) SetSchema(schema string) {
	db.schema = schema
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
	db := app.Default().DbManager().Db("{{.Database}}")

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
