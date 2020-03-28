package appy

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

const (
	LOGGER_DB_PREFIX = "[DB] "
)

var (
	dbMigratePath = "db/migrate/"
)

type DBer interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	Close() error
	Config() *DBConfig
	Connect() error
	ConnectDB(database string) error
	CreateDB(database string) error
	DropDB(database string) error
	DumpSchema(database string) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GenerateMigration(name, target string, tx bool) error
	Migrate() error
	MigrateStatus() ([][]string, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	RegisterMigration(up func(*DB) error, down func(*DB) error, args ...string)
	RegisterMigrationTx(upTx func(*DBTx) error, downTx func(*DBTx) error, args ...string)
	RegisterSeedTx(seed func(*DBTx) error)
	Rollback() error
	Seed() error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Schema() string
	SetSchema(schema string)
}

type DB struct {
	*sqlx.DB
	config     *DBConfig
	logger     *Logger
	migrations []*DBMigration
	mu         *sync.Mutex
	schema     string
	seed       func(*DBTx) error
	support    Supporter
}

type DBTx = sql.Tx

// NewDB initializes the database handler that is used to connect to the database.
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

// BindNamed binds a query using the DB driver's bindvar type.
func (db *DB) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, arg)
	return db.DB.BindNamed(query, arg)
}

// Config returns the database config.
func (db *DB) Config() *DBConfig {
	return db.config
}

// Connect establishes a connection to the database specified in URI and assign the database
// handler which is safe for concurrent use by multiple goroutines and maintains its own
// connection pool.
func (db *DB) Connect() error {
	wrapper, err := sqlx.Connect(db.config.Adapter, db.config.URI)
	if err != nil {
		return err
	}

	err = db.setupWrapper(wrapper)
	if err != nil {
		return err
	}

	return nil
}

// ConnectDB establishes a connection to the specific database and assign the database handler
// which is safe for concurrent use by multiple goroutines and maintains its own connection pool.
func (db *DB) ConnectDB(database string) error {
	uri := db.config.URI
	if database != "" {
		u, err := url.Parse(uri)
		if err != nil {
			return err
		}

		switch db.config.Adapter {
		case "mysql":
			uri = strings.ReplaceAll(u.String(), "/"+db.config.Database, "/"+database)
		case "postgres":
			u.Path = "/" + database
			uri = u.String()
		case "sqlite3":
		}
	}

	wrapper, err := sqlx.Connect(db.config.Adapter, uri)
	if err != nil {
		return err
	}

	err = db.setupWrapper(wrapper)
	if err != nil {
		return err
	}

	return nil
}

// Create creates the database.
func (db *DB) CreateDB(database string) error {
	switch db.config.Adapter {
	case "mysql", "postgres":
		_, err := db.Exec("CREATE DATABASE " + database)
		if err != nil {
			return err
		}
	case "sqlite3":
		_, err := os.Create(database)
		if err != nil {
			return err
		}
	}

	return nil
}

// Drop drops the database.
func (db *DB) DropDB(database string) error {
	if db.config.Adapter == "postgres" {
		_, err := db.Exec(
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`,
			database,
		)

		if err != nil {
			return err
		}
	}

	switch db.config.Adapter {
	case "mysql", "postgres":
		_, err := db.Exec("DROP DATABASE " + database)
		if err != nil {
			return err
		}
	case "sqlite3":
		err := os.Remove(database)
		if err != nil {
			return err
		}
	}

	return nil
}

// DumpSchema dumps the database schema into "db/migrate/<database>/schema.go".
func (db *DB) DumpSchema(database string) error {
	path := dbMigratePath + database
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	err = db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	var (
		outBytes    bytes.Buffer
		out         string
		versionRows *sql.Rows
		versions    []string
	)

	switch db.config.Adapter {
	case "mysql":
		_, err = exec.LookPath("mysqldump")
		if err != nil {
			return err
		}

		dumpArgs := []string{
			"--no-data",
			"--routines",
			"--skip-comments",
			"--skip-quote-names",
			"--host", db.config.Host,
			"--port", db.config.Port,
			"--user", db.config.Username,
			db.config.Database,
		}
		dumpCmd := exec.Command("mysqldump", dumpArgs...)
		dumpCmd.Env = os.Environ()
		dumpCmd.Env = append(dumpCmd.Env, []string{"MYSQL_PWD=" + db.config.Password}...)
		dumpCmd.Stdout = &outBytes
		dumpCmd.Stderr = os.Stderr

		err = dumpCmd.Run()
		if err != nil {
			return err
		}

		out = outBytes.String()
		out = strings.Trim(out, "\n")

		versionRows, err = db.Query(
			fmt.Sprintf(
				"SELECT version FROM %s.%s ORDER BY version ASC",
				db.config.Database,
				db.config.SchemaMigrationsTable,
			),
		)
	case "postgres":
		_, err = exec.LookPath("pg_dump")
		if err != nil {
			return err
		}

		dumpArgs := []string{
			"-s", "-x", "-O", "--no-comments",
			"-d", db.config.Database,
			"-n", db.config.SchemaSearchPath,
			"-h", db.config.Host,
			"-p", db.config.Port,
			"-U", db.config.Username,
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

		versionRows, err = db.Query(
			fmt.Sprintf(
				"SELECT version FROM %s.%s ORDER BY version ASC",
				db.config.SchemaSearchPath,
				db.config.SchemaMigrationsTable,
			),
		)
	}

	if err != nil {
		return err
	}

	for versionRows.Next() {
		var version string
		err = versionRows.Scan(&version)
		if err != nil {
			return err
		}
		versions = append(versions, version)
	}
	versionRows.Close()

	if len(versions) > 0 {
		out += fmt.Sprintf("\n\nINSERT INTO %s.%s (version) VALUES\n", db.config.SchemaSearchPath, db.config.SchemaMigrationsTable)

		for idx, version := range versions {
			out += "('" + version + "')"

			if idx == len(versions)-1 {
				out += ";\n"
			} else {
				out += ",\n"
			}
		}
	}

	tpl, err := schemaDumpTpl(database, out)
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
	path := dbMigratePath + target
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

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		err = tx.Commit()
		if err != nil {
			defer tx.Rollback()
			return err
		}
	}

	for _, m := range db.migrations {
		if !db.support.ArrayContains(migratedVersions, m.Version) {
			if m.UpTx != nil {
				err = m.UpTx(tx)
				if err != nil {
					defer tx.Rollback()
					return err
				}

				err = db.addSchemaMigration(tx, m)
				if err != nil {
					defer tx.Rollback()
					return err
				}

				err = tx.Commit()
				if err != nil {
					defer tx.Rollback()
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

	var migrationStatus [][]string
	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		err = tx.Commit()
		if err != nil {
			defer tx.Rollback()
			return nil, err
		}
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
		defer tx.Rollback()
		return nil, err
	}

	return migrationStatus, nil
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters
// in the query.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.Exec(query, args...)
}

// ExecContext executes a query without returning any rows. The args are for any placeholder
// parameters in the query.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder
// parameters in the query.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.Query(query, args...)
}

// QueryContext executes a query that returns rows, typically a SELECT. The args are for any
// placeholder parameters in the query.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always returns a
// non-nil value. Errors are deferred until Row's Scan method is called.
//
// If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan
// scans the first selected row and discards the rest.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext
// always returns a non-nil value. Errors are deferred until Row's Scan method is called.
//
// If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan
// scans the first selected row and discards the rest.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryRowContext(ctx, query, args...)
}

// QueryRowx queries the database and returns an *sqlx.Row. Any placeholder parameters are replaced
// with supplied args.
func (db *DB) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryRowx(query, args...)
}

// QueryRowxContext queries the database and returns an *sqlx.Row. Any placeholder parameters are
// replaced with supplied args.
func (db *DB) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryRowxContext(ctx, query, args...)
}

// Queryx queries the database and returns an *sqlx.Rows. Any placeholder parameters are replaced
// with supplied args.
func (db *DB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.Queryx(query, args...)
}

// QueryxContext queries the database and returns an *sqlx.Rows. Any placeholder parameters are
// replaced with supplied args.
func (db *DB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.QueryxContext(ctx, query, args...)
}

// RegisterMigration registers the up/down migrations that won't be executed in transaction.
func (db *DB) RegisterMigration(up func(*DB) error, down func(*DB) error, args ...string) {
	err := db.registerMigration(up, down, nil, nil, args...)
	if err != nil {
		db.logger.Fatal(err)
	}
}

// RegisterMigrationTx registers the up/down migrations that will be executed in transaction.
func (db *DB) RegisterMigrationTx(upTx func(*DBTx) error, downTx func(*DBTx) error, args ...string) {
	err := db.registerMigration(nil, nil, upTx, downTx, args...)
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

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		err = tx.Commit()
		if err != nil {
			defer tx.Rollback()
			return err
		}
	}

	if len(migratedVersions) > 0 {
		for i := len(db.migrations) - 1; i > -1; i-- {
			m := db.migrations[i]

			if migratedVersions[len(migratedVersions)-1] == m.Version {
				if m.DownTx != nil {
					err = m.DownTx(tx)
					if err != nil {
						defer tx.Rollback()
						return err
					}

					err = db.removeSchemaMigration(tx, m)
					if err != nil {
						defer tx.Rollback()
						return err
					}

					err = tx.Commit()
					if err != nil {
						defer tx.Rollback()
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

	return nil
}

// Select using this DB. Any placeholder parameters are replaced with supplied args.
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.Select(dest, query, args...)
}

// SelectContext using this DB. Any placeholder parameters are replaced with supplied args.
func (db *DB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(LOGGER_DB_PREFIX+query, args...)
	return db.DB.SelectContext(ctx, dest, query, args...)
}

// Schema returns the database schema.
func (db *DB) Schema() string {
	return db.schema
}

// SetSchema sets the database schema.
func (db *DB) SetSchema(schema string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.schema = schema
}

// Seed runs the seeding for the current environment.
func (db *DB) Seed() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if db.seed != nil {
		err := db.seed(tx)
		if err != nil {
			defer tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		defer tx.Rollback()
		return err
	}

	return nil
}

func (db *DB) addMigration(migration *DBMigration) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.migrations = append(db.migrations, migration)
}

func (db *DB) addSchemaMigration(tx *DBTx, migration *DBMigration) error {
	var query string

	switch db.config.Adapter {
	case "mysql", "sqlite":
		query = fmt.Sprintf(
			"INSERT INTO %s.%s (version) VALUES (%s)",
			db.config.Database,
			db.config.SchemaMigrationsTable,
			migration.Version,
		)

		if tx != nil {
			_, err := tx.Exec(query)
			return err
		}
	case "postgres":
		query = fmt.Sprintf(
			"INSERT INTO %s.%s (version) VALUES (%s)",
			db.config.SchemaSearchPath,
			db.config.SchemaMigrationsTable,
			migration.Version,
		)

		if tx != nil {
			_, err := tx.Exec(query)
			return err
		}
	}

	_, err := db.Exec(query)
	return err
}

func (db *DB) removeSchemaMigration(tx *DBTx, migration *DBMigration) error {
	var query string

	switch db.config.Adapter {
	case "mysql", "sqlite":
		query = fmt.Sprintf(
			`DELETE FROM %s.%s WHERE version = '%s'`,
			db.config.Database,
			db.config.SchemaMigrationsTable,
			migration.Version,
		)

		if tx != nil {
			_, err := tx.Exec(query)
			return err
		}
	case "postgres":
		query = fmt.Sprintf(
			`DELETE FROM %s.%s WHERE version = '%s'`,
			db.config.SchemaSearchPath,
			db.config.SchemaMigrationsTable,
			migration.Version,
		)

		if tx != nil {
			_, err := tx.Exec(query)
			return err
		}
	}

	_, err := db.Exec(query)
	return err
}

func (db *DB) ensureSchemaMigrationsTable() error {
	var (
		count int
		err   error
		rows  *sql.Rows
	)

	switch db.config.Adapter {
	case "mysql":
		rows, err = db.Query(
			`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`,
			db.config.Database,
			db.config.SchemaMigrationsTable,
		)
	case "postgres":
		rows, err = db.Query(
			`SELECT COUNT(*) FROM pg_tables WHERE schemaname = $1 AND tablename = $2`,
			db.config.SchemaSearchPath,
			db.config.SchemaMigrationsTable,
		)
	case "sqlite3":
		rows, err = db.Query(
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?`,
			db.config.SchemaMigrationsTable,
		)
	}

	if err != nil {
		return err
	}

	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		return err
	}
	rows.Close()

	if count < 1 {
		switch db.config.Adapter {
		case "mysql", "sqlite3":
			_, err = db.Exec("CREATE TABLE IF NOT EXISTS " + db.config.SchemaMigrationsTable + " (`version` varchar(64), PRIMARY KEY (`version`)) ")
			if err != nil {
				return err
			}
		case "postgres":
			_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS " + db.config.SchemaSearchPath)
			if err != nil {
				return err
			}

			_, err = db.Exec("CREATE TABLE " + db.config.SchemaMigrationsTable + " (version VARCHAR PRIMARY KEY) ")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *DB) migratedVersions(tx *DBTx) ([]string, error) {
	var (
		err  error
		rows *sql.Rows
	)

	switch db.config.Adapter {
	case "mysql", "sqlite3":
		rows, err = tx.Query(fmt.Sprintf("SELECT version FROM %s.%s ORDER BY version ASC", db.config.Database, db.config.SchemaMigrationsTable))
	case "postgres":
		rows, err = tx.Query(fmt.Sprintf("SELECT version FROM %s.%s ORDER BY version ASC", db.config.SchemaSearchPath, db.config.SchemaMigrationsTable))
	}

	if err != nil {
		return nil, err
	}

	migratedVersions := []string{}
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}

		migratedVersions = append(migratedVersions, version)
	}
	rows.Close()

	return migratedVersions, nil
}

func (db *DB) registerMigration(up func(*DB) error, down func(*DB) error, upTx func(*DBTx) error, downTx func(*DBTx) error, args ...string) error {
	file := migrationFile()

	if len(args) > 0 {
		file = args[0]
	}

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

func (db *DB) setupWrapper(wrapper *sqlx.DB) error {
	db.DB = wrapper
	db.DB.SetConnMaxLifetime(db.config.ConnMaxLifetime)
	db.DB.SetMaxIdleConns(db.config.MaxIdleConns)
	db.DB.SetMaxOpenConns(db.config.MaxOpenConns)

	return db.DB.Ping()
}
