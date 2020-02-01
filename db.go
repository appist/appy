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
	// DB is a database handle representing a pool of zero or more underlying connections. It's safe for concurrent use by
	// multiple goroutines.
	DB struct {
		*pg.DB
		config     *DBConfig
		logger     *Logger
		migrations []*DBMigration
		mu         *sync.Mutex
		schema     string
		seed       func(*DBTx) error
		support    Supporter
	}

	// DBMigration contains database migration.
	DBMigration struct {
		File    string
		Version string
		Down    func(*DB) error
		DownTx  func(*DBTx) error
		Up      func(*DB) error
		UpTx    func(*DBTx) error
	}

	// DBQueryEvent keeps the query event information.
	DBQueryEvent = pg.QueryEvent

	// SchemaMigration is a model that maps to `schema_migrations` table.
	SchemaMigration struct {
		Version string
	}

	// DBTx is an in-progress database transaction. It is safe for concurrent use by multiple goroutines.
	DBTx = pg.Tx
)

var (
	// DBArray accepts a slice and returns a wrapper for working with PostgreSQL array data type.
	// For struct fields you can use array tag:
	//
	//    Emails  []string `pg:",array"`
	DBArray = pg.Array

	// DBHStore accepts a map and returns a wrapper for working with hstore data type.
	// For struct fields you can use hstore tag:
	//
	//    Attrs map[string]string `pg:",hstore"`
	DBHStore = pg.Hstore

	// DBSafeQuery replaces any placeholders found in the query.
	DBSafeQuery = pg.SafeQuery

	// DBScan returns ColumnScanner that copies the columns in the row into the values.
	DBScan = pg.Scan
)

// NewDB initializes the DB handler that is used to connect to the database.
func NewDB(config *DBConfig, logger *Logger, support Supporter) *DB {
	return &DB{
		nil,
		config,
		logger,
		nil,
		&sync.Mutex{},
		"",
		nil,
		support,
	}
}

// Config returns the database config.
func (db *DB) Config() *DBConfig {
	return db.config
}

// Connect establishes a connection to the database using provided options and assign the database Handler which is
// safe for concurrent use by multiple goroutines and maintains its own connection pool.
func (db *DB) Connect() error {
	opts := db.config.Options
	db.DB = pg.Connect(&opts)
	db.AddQueryHook(db.logger)

	_, err := db.Exec("SELECT 1 /* appy framework */")
	return err
}

// Create creates the database.
func (db *DB) Create() []error {
	var errs []error
	_, err := db.Exec(`CREATE DATABASE ?`, DBSafeQuery(db.config.Database))
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.Exec(`CREATE DATABASE ?`, DBSafeQuery(db.config.Database+"_test"))
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Drop creates the database.
func (db *DB) Drop() []error {
	var errs []error
	_, err := db.Exec(
		`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '?'`,
		DBSafeQuery(db.config.Database),
	)
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.Exec(`DROP DATABASE ?`, DBSafeQuery(db.config.Database))
	if err != nil {
		errs = append(errs, err)
	}

	_, err = db.Exec(`DROP DATABASE ?`, DBSafeQuery(db.config.Database+"_test"))
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

// DumpSchema uses `pg_dump` to dump the database schema.
func (db *DB) DumpSchema(name string) error {
	var (
		outBytes bytes.Buffer
		out      string
	)

	path := "db/migrate/" + name
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

	err = dumpCmd.Run()
	if err != nil {
		return err
	}

	out = outBytes.String()
	out = regexp.MustCompile(`(?i)--\n-- postgresql database dump.*\n--\n\n`).ReplaceAllString(out, "")
	out = regexp.MustCompile(`(?i)--\ dumped.*\n(\n)?`).ReplaceAllString(out, "")
	out = regexp.MustCompile(`(?i)create\ extension`).ReplaceAllString(out, "CREATE EXTENSION IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ schema`).ReplaceAllString(out, "CREATE SCHEMA IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ sequence`).ReplaceAllString(out, "CREATE SEQUENCE IF NOT EXISTS")
	out = regexp.MustCompile(`(?i)create\ table`).ReplaceAllString(out, "CREATE TABLE IF NOT EXISTS")
	out = strings.Trim(out, "\n")

	var schemaMigrations []SchemaMigration
	_, err = db.Query(
		&schemaMigrations,
		`SELECT version FROM ?.? ORDER BY version ASC`,
		DBSafeQuery(db.config.SchemaSearchPath),
		DBSafeQuery(db.config.SchemaMigrationsTable),
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

// GenerateMigration generates the migration file for the target database.
func (db *DB) GenerateMigration(name, target string, tx bool) error {
	path := "db/migrate/" + target
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	ts := time.Now()
	fn := path + "/" + ts.Format("20060102150405") + "_" + db.support.ToSnakeCase(name) + ".go"
	db.logger.Infof("Generating migration '%s' for '%s' database...", fn, target)

	tpl, err := migrationTpl(target, tx)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fn, tpl, 0644)
	if err != nil {
		return err
	}

	db.logger.Infof("Generating migration '%s' for '%s' database... DONE", fn, target)
	return nil
}

// Migrate runs migrations for the current environment that have not run yet.
func (db *DB) Migrate() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err = tx.Close(); err != nil {
			db.logger.Fatal(err)
		}
	}()

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return err
	}

	for _, m := range db.migrations {
		if !db.support.ArrayContains(migratedVersions, m.Version) {
			if m.UpTx != nil {
				err = m.UpTx(tx)
				if err != nil {
					_ = tx.Rollback()
					return err
				}

				err = db.addSchemaMigration(tx, m)
				if err != nil {
					_ = tx.Rollback()
					return err
				}

				continue
			}

			err = m.Up(db)
			if err != nil {
				return err
			}

			err = db.addSchemaMigration(nil, m)
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
func (db *DB) MigrateStatus() ([][]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = tx.Close(); err != nil {
			db.logger.Fatal(err)
		}
	}()

	var migrationStatus [][]string
	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return nil, err
	}

	wd, _ := os.Getwd()
	for _, m := range db.migrations {
		status := "down"
		if db.support.ArrayContains(migratedVersions, m.Version) {
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

// RegisterMigration registers the up/down migrations that won't be executed in transaction.
func (db *DB) RegisterMigration(up func(*DB) error, down func(*DB) error) {
	err := db.registerMigration(up, down, nil, nil)
	if err != nil {
		db.logger.Fatal(err)
	}
}

// RegisterMigrationTx registers the up/down migrations that will be executed in transaction.
func (db *DB) RegisterMigrationTx(upTx func(*DBTx) error, downTx func(*DBTx) error) {
	err := db.registerMigration(nil, nil, upTx, downTx)
	if err != nil {
		db.logger.Fatal(err)
	}
}

// RegisterSeedTx registers the seeding that will be executed in transaction.
func (db *DB) RegisterSeedTx(seed func(*DBTx) error) {
	db.seed = seed
}

// Rollback rolls back the last migration for the current environment.
func (db *DB) Rollback() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err = tx.Close(); err != nil {
			db.logger.Fatal(err)
		}
	}()

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
						_ = tx.Rollback()
						return err
					}

					err = db.removeSchemaMigration(tx, m)
					if err != nil {
						_ = tx.Rollback()
						return err
					}

					continue
				}

				err = m.Down(db)
				if err != nil {
					return err
				}

				err = db.removeSchemaMigration(nil, m)
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

// Schema returns the database schema.
func (db *DB) Schema() string {
	return db.schema
}

// SetSchema stores the database schema.
func (db *DB) SetSchema(schema string) {
	db.schema = schema
}

// Seed runs the seeding for the current environment.
func (db *DB) Seed() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err = tx.Close(); err != nil {
			db.logger.Fatal(err)
		}
	}()

	if db.seed != nil {
		err := db.seed(tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) addMigration(newMigration *DBMigration) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.migrations = append(db.migrations, newMigration)
}

func (db *DB) addSchemaMigration(tx *DBTx, targetMigration *DBMigration) error {
	query := `INSERT INTO ?.? (version) VALUES (?)`
	if tx != nil {
		_, err := tx.Exec(
			query,
			DBSafeQuery(db.config.SchemaSearchPath),
			DBSafeQuery(db.config.SchemaMigrationsTable),
			DBSafeQuery(targetMigration.Version),
		)

		return err
	}

	_, err := db.Exec(
		query,
		DBSafeQuery(db.config.SchemaSearchPath),
		DBSafeQuery(db.config.SchemaMigrationsTable),
		DBSafeQuery(targetMigration.Version),
	)

	return err
}

func (db *DB) ensureSchemaMigrationsTable() error {
	count, err := db.
		Model().
		Table("pg_tables").
		Where("schemaname = '?'", DBSafeQuery(db.config.SchemaSearchPath)).
		Where("tablename = '?'", DBSafeQuery(db.config.SchemaMigrationsTable)).
		Count()

	if err != nil {
		return err
	}

	if count < 1 {
		_, err = db.Exec(`CREATE SCHEMA IF NOT EXISTS ?`, DBSafeQuery(db.config.SchemaSearchPath))
		if err != nil {
			return err
		}

		_, err = db.Exec(`CREATE TABLE ? (version VARCHAR PRIMARY KEY)`, DBSafeQuery(db.config.SchemaMigrationsTable))
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) migratedVersions(tx *DBTx) ([]string, error) {
	var schemaMigrations []SchemaMigration
	_, err := tx.Query(
		&schemaMigrations,
		`SELECT version FROM ?.? ORDER BY version ASC`,
		DBSafeQuery(db.config.SchemaSearchPath),
		DBSafeQuery(db.config.SchemaMigrationsTable),
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

func (db *DB) registerMigration(up func(*DB) error, down func(*DB) error, upTx func(*DBTx) error, downTx func(*DBTx) error) error {
	file := migrationFile()
	version, err := migrationVersion(file)
	if err != nil {
		return err
	}

	db.addMigration(&DBMigration{
		File:    file,
		Version: version,
		Down:    down,
		DownTx:  downTx,
		Up:      up,
		UpTx:    upTx,
	})

	return nil
}

func (db *DB) removeSchemaMigration(tx *DBTx, targetMigration *DBMigration) error {
	query := `DELETE FROM ?.? WHERE version = '?'`
	if tx != nil {
		_, err := tx.Exec(
			query,
			DBSafeQuery(db.config.SchemaSearchPath),
			DBSafeQuery(db.config.SchemaMigrationsTable),
			DBSafeQuery(targetMigration.Version),
		)

		return err
	}

	_, err := db.Exec(
		query,
		DBSafeQuery(db.config.SchemaSearchPath),
		DBSafeQuery(db.config.SchemaMigrationsTable),
		DBSafeQuery(targetMigration.Version),
	)

	return err
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

func migrationTpl(database string, tx bool) ([]byte, error) {
	type data struct {
		Database, Module string
		Tx               bool
	}

	t, err := template.New("migration").Parse(
		`package {{.Database}}

import (
	"github.com/appist/appy"

	"{{.Module}}/pkg/app"
)

func init() {
	db := app.DBManager.DB("{{.Database}}")

	if db != nil {
		db.RegisterMigration{{if .Tx}}Tx{{end}}(
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
		Tx:       tx,
	})

	if err != nil {
		return []byte(""), err
	}

	return tpl.Bytes(), err
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
	db := app.DBManager.DB("{{.Database}}")

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

func seedTpl(database, schema string) ([]byte, error) {
	type data struct {
		Database, Module string
	}

	t, err := template.New("seed").Parse(
		`package {{.Database}}

import (
	"github.com/appist/appy"

	"{{.Module}}/pkg/app"
)

func init() {
	db := app.DBManager.DB("{{.Database}}")

	if db != nil {
		db.RegisterSeedTx(
			func(db *appy.DBTx) error {
				return nil
			},
		)
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
