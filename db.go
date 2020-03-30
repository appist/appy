package appy

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	// Automatically import mysql driver to make it easier for appy's users.
	_ "github.com/go-sql-driver/mysql"

	// Automatically import postgres driver to make it easier for appy's users.
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

const (
	loggerDBPrefix = "[DB] "
)

var (
	dbMigratePath = "db/migrate/"
)

type (
	// DBer implements all DB methods and is useful for mocking DB in unit tests.
	DBer interface {
		Begin() (*DBTx, error)
		BeginContext(ctx context.Context, opts *sql.TxOptions) (*DBTx, error)
		Close() error
		Config() *DBConfig
		Conn(ctx context.Context) (*sql.Conn, error)
		Connect() error
		ConnectDB(database string) error
		CreateDB(database string) error
		Driver() driver.Driver
		DriverName() string
		DropDB(database string) error
		DumpSchema(database string) error
		Exec(query string, args ...interface{}) (sql.Result, error)
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		GenerateMigration(name, target string, tx bool) error
		Get(dest interface{}, query string, args ...interface{}) error
		GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		Migrate() error
		MigrateStatus() ([][]string, error)
		NamedExec(query string, arg interface{}) (sql.Result, error)
		NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
		NamedQuery(query string, arg interface{}) (*DBRows, error)
		NamedQueryContext(ctx context.Context, query string, arg interface{}) (*DBRows, error)
		Ping() error
		PingContext(ctx context.Context) error
		Prepare(query string) (*DBStmt, error)
		PrepareContext(ctx context.Context, query string) (*DBStmt, error)
		PrepareNamed(query string) (*DBNamedStmt, error)
		PrepareNamedContext(ctx context.Context, query string) (*DBNamedStmt, error)
		Query(query string, args ...interface{}) (*DBRows, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*DBRows, error)
		QueryRow(query string, args ...interface{}) *DBRow
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *DBRow
		Rebind(query string) string
		RegisterMigration(up func(*DB) error, down func(*DB) error, args ...string) error
		RegisterMigrationTx(upTx func(*DBTx) error, downTx func(*DBTx) error, args ...string) error
		RegisterSeedTx(seed func(*DBTx) error)
		Rollback() error
		Schema() string
		Seed() error
		Select(dest interface{}, query string, args ...interface{}) error
		SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		SetConnMaxLifetime(d time.Duration)
		SetMaxIdleConns(n int)
		SetMaxOpenConns(n int)
		SetSchema(schema string)
		Stats() sql.DBStats
	}

	// DB manages the database config/connection/migrations.
	DB struct {
		*sqlx.DB
		config     *DBConfig
		logger     *Logger
		migrations []*DBMigration
		mu         *sync.Mutex
		schema     string
		seed       func(*DBTx) error
		support    Supporter
	}

	// DBRow is a wrapper around sqlx.Row.
	DBRow struct {
		*sqlx.Row
	}

	// DBRows is a wrapper around sqlx.Rows.
	DBRows struct {
		*sqlx.Rows
	}

	// DBNamedStmt is a prepared statement that executes named queries.  Prepare it how you would
	// execute a NamedQuery, but pass in a struct or map when executing.
	DBNamedStmt struct {
		*sqlx.NamedStmt
	}
)

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

// Begin starts a transaction. The default isolation level is dependent on the driver.
func (db *DB) Begin() (*DBTx, error) {
	db.logger.Info(formatDBQuery("BEGIN;"))

	tx, err := db.DB.Beginx()
	return &DBTx{tx, db.logger}, err
}

// BeginContext starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back. If the context
// is canceled, the sql package will roll back the transaction. Tx.Commit will return an error if
// the context provided to BeginContext is canceled.
//
// The provided TxOptions is optional and may be nil if defaults should be used. If a non-default
// isolation level is used that the driver doesn't support, an error will be returned.
func (db *DB) BeginContext(ctx context.Context, opts *sql.TxOptions) (*DBTx, error) {
	db.logger.Info(formatDBQuery("BEGIN;"))

	tx, err := db.DB.BeginTxx(ctx, opts)
	return &DBTx{tx, db.logger}, err
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

// CreateDB creates the database.
func (db *DB) CreateDB(database string) error {
	_, err := db.Exec(
		fmt.Sprintf(
			"CREATE DATABASE %s;",
			database,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

// DropDB drops the database.
func (db *DB) DropDB(database string) error {
	if db.config.Adapter == "postgres" {
		_, err := db.Exec(
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1;`,
			database,
		)

		if err != nil {
			return err
		}
	}

	_, err := db.Exec(
		fmt.Sprintf(
			"DROP DATABASE %s;",
			database,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

// DumpSchema dumps the database schema into "db/migrate/<dbname>/schema.go".
func (db *DB) DumpSchema(dbname string) error {
	path := dbMigratePath + dbname
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	err = db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	var (
		outBytes      bytes.Buffer
		database, out string
		versionRows   *DBRows
		versions      []string
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
				"SELECT version FROM %s.%s ORDER BY version ASC;",
				db.config.Database,
				db.config.SchemaMigrationsTable,
			),
		)
		database = db.config.Database
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
				"SELECT version FROM %s.%s ORDER BY version ASC;",
				db.config.SchemaSearchPath,
				db.config.SchemaMigrationsTable,
			),
		)
		database = db.config.SchemaSearchPath
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
		out += fmt.Sprintf("\n\nINSERT INTO %s.%s (version) VALUES\n", database, db.config.SchemaMigrationsTable)

		for idx, version := range versions {
			out += "('" + version + "')"

			if idx == len(versions)-1 {
				out += ";\n"
			} else {
				out += ",\n"
			}
		}
	}

	out = strings.Trim(out, "\n")
	tpl, err := schemaDumpTpl(dbname, out)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path+"/schema.go", tpl, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters
// in the query.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db.logger.Infof(formatDBQuery(query), args...)
	return db.DB.Exec(query, args...)
}

// ExecContext executes a query without returning any rows. The args are for any placeholder
// parameters in the query.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db.logger.Infof(formatDBQuery(query), args...)
	return db.DB.ExecContext(ctx, query, args...)
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

// Get using this DB. Any placeholder parameters are replaced with supplied args. An error is
// returned if the result set is empty.
func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(formatDBQuery(query), args...)
	return db.DB.Get(dest, query, args...)
}

// GetContext using this DB. Any placeholder parameters are replaced with supplied args. An error
// is returned if the result set is empty.
func (db *DB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(formatDBQuery(query), args...)
	return db.DB.GetContext(ctx, dest, query, args...)
}

// Migrate runs migrations for the current environment that have not run yet.
func (db *DB) Migrate() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	migratedVersions, err := db.migratedVersions()
	if err != nil {
		return err
	}

	for _, m := range db.migrations {
		if !db.support.ArrayContains(migratedVersions, m.Version) {
			if m.UpTx != nil {
				tx, err := db.Begin()
				if err != nil {
					return err
				}

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

	var migrationStatus [][]string
	migratedVersions, err := db.migratedVersions()
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

	return migrationStatus, nil
}

// NamedExec using this DB. Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedExec(query string, arg interface{}) (sql.Result, error) {
	db.logger.Infof(formatDBQuery(query), arg)
	return db.DB.NamedExec(query, arg)
}

// NamedExecContext using this DB. Any named placeholder parameters are replaced with fields from
// arg.
func (db *DB) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	db.logger.Infof(formatDBQuery(query), arg)
	return db.DB.NamedExecContext(ctx, query, arg)
}

// NamedQuery using this DB. Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedQuery(query string, arg interface{}) (*DBRows, error) {
	db.logger.Infof(formatDBQuery(query), arg)

	rows, err := db.DB.NamedQuery(query, arg)
	return &DBRows{rows}, err
}

// NamedQueryContext using this DB. Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*DBRows, error) {
	db.logger.Infof(formatDBQuery(query), arg)

	rows, err := db.DB.NamedQueryContext(ctx, query, arg)
	return &DBRows{rows}, err
}

// PrepareNamed returns a DBNamedStmt.
func (db *DB) PrepareNamed(query string) (*DBNamedStmt, error) {
	db.logger.Info(formatDBQuery(query))

	namedStmt, err := db.DB.PrepareNamed(query)
	return &DBNamedStmt{namedStmt}, err
}

// PrepareNamedContext returns DBNamedStmt.
func (db *DB) PrepareNamedContext(ctx context.Context, query string) (*DBNamedStmt, error) {
	db.logger.Info(formatDBQuery(query))

	namedStmt, err := db.DB.PrepareNamedContext(ctx, query)
	return &DBNamedStmt{namedStmt}, err
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder
// parameters in the query.
func (db *DB) Query(query string, args ...interface{}) (*DBRows, error) {
	db.logger.Infof(formatDBQuery(query), args...)

	rows, err := db.DB.Queryx(query, args...)
	return &DBRows{rows}, err
}

// QueryContext executes a query that returns rows, typically a SELECT. The args are for any
// placeholder parameters in the query.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*DBRows, error) {
	db.logger.Infof(formatDBQuery(query), args...)

	rows, err := db.DB.QueryxContext(ctx, query, args...)
	return &DBRows{rows}, err
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always returns a
// non-nil value. Errors are deferred until Row's Scan method is called.
//
// If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan
// scans the first selected row and discards the rest.
func (db *DB) QueryRow(query string, args ...interface{}) *DBRow {
	db.logger.Infof(formatDBQuery(query), args...)

	row := db.DB.QueryRowx(query, args...)
	return &DBRow{row}
}

// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext
// always returns a non-nil value. Errors are deferred until Row's Scan method is called.
//
// If the query selects no rows, the *Row's Scan will return ErrNoRows. Otherwise, the *Row's Scan
// scans the first selected row and discards the rest.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *DBRow {
	db.logger.Infof(formatDBQuery(query), args...)

	row := db.DB.QueryRowxContext(ctx, query, args...)
	return &DBRow{row}
}

// RegisterMigration registers the up/down migrations that won't be executed in transaction.
func (db *DB) RegisterMigration(up func(*DB) error, down func(*DB) error, args ...string) error {
	err := db.registerMigration(up, down, nil, nil, args...)
	if err != nil {
		return err
	}

	return nil
}

// RegisterMigrationTx registers the up/down migrations that will be executed in transaction.
func (db *DB) RegisterMigrationTx(upTx func(*DBTx) error, downTx func(*DBTx) error, args ...string) error {
	err := db.registerMigration(nil, nil, upTx, downTx, args...)
	if err != nil {
		return err
	}

	return nil
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

	migratedVersions, err := db.migratedVersions()
	if err != nil {
		return err
	}

	if len(migratedVersions) > 0 {
		for i := len(db.migrations) - 1; i > -1; i-- {
			m := db.migrations[i]

			if migratedVersions[len(migratedVersions)-1] == m.Version {
				if m.DownTx != nil {
					tx, err := db.Begin()
					if err != nil {
						return err
					}

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

// Prepare creates a prepared statement for later queries or executions. Multiple queries or
// executions may be run concurrently from the returned statement. The caller must call the
// statement's Close method when the statement is no longer needed.
func (db *DB) Prepare(query string) (*DBStmt, error) {
	stmt, err := db.DB.Preparex(query)
	return &DBStmt{stmt, db.logger, query}, err
}

// PrepareContext creates a prepared statement for later queries or executions. Multiple queries
// or executions may be run concurrently from the returned statement. The caller must call the
// statement's Close method when the statement is no longer needed.
//
// The provided context is used for the preparation of the statement, not for the execution of
// the statement.
func (db *DB) PrepareContext(ctx context.Context, query string) (*DBStmt, error) {
	stmt, err := db.DB.PreparexContext(ctx, query)
	return &DBStmt{stmt, db.logger, query}, err
}

// Select using this DB. Any placeholder parameters are replaced with supplied args.
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(formatDBQuery(query), args...)
	return db.DB.Select(dest, query, args...)
}

// SelectContext using this DB. Any placeholder parameters are replaced with supplied args.
func (db *DB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	db.logger.Infof(formatDBQuery(query), args...)
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
	case "mysql":
		query = fmt.Sprintf(
			"INSERT INTO %s.%s (version) VALUES (%s);",
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
			"INSERT INTO %s.%s (version) VALUES (%s);",
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
	case "mysql":
		query = fmt.Sprintf(
			`DELETE FROM %s.%s WHERE version = '%s';`,
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
			`DELETE FROM %s.%s WHERE version = '%s';`,
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
	)

	switch db.config.Adapter {
	case "mysql":
		err = db.Get(&count,
			`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?;`,
			db.config.Database,
			db.config.SchemaMigrationsTable,
		)
	case "postgres":
		err = db.Get(&count,
			`SELECT COUNT(*) FROM pg_tables WHERE schemaname = $1 AND tablename = $2;`,
			db.config.SchemaSearchPath,
			db.config.SchemaMigrationsTable,
		)
	}

	if err != nil {
		return err
	}

	if count < 1 {
		switch db.config.Adapter {
		case "mysql":
			_, err = db.Exec(
				fmt.Sprintf(
					"CREATE TABLE IF NOT EXISTS %s (version varchar(64), PRIMARY KEY (version));",
					db.config.SchemaMigrationsTable,
				),
			)
			if err != nil {
				return err
			}
		case "postgres":
			_, err = db.Exec(
				fmt.Sprintf(
					"CREATE SCHEMA IF NOT EXISTS %s;",
					db.config.SchemaSearchPath,
				),
			)
			if err != nil {
				return err
			}

			_, err = db.Exec(
				fmt.Sprintf(
					"CREATE TABLE %s (version VARCHAR PRIMARY KEY);",
					db.config.SchemaMigrationsTable,
				),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *DB) migratedVersions() ([]string, error) {
	var (
		err  error
		rows *DBRows
	)

	switch db.config.Adapter {
	case "mysql":
		rows, err = db.Query(
			fmt.Sprintf(
				"SELECT version FROM %s.%s ORDER BY version ASC;",
				db.config.Database,
				db.config.SchemaMigrationsTable,
			),
		)
	case "postgres":
		rows, err = db.Query(
			fmt.Sprintf(
				"SELECT version FROM %s.%s ORDER BY version ASC;",
				db.config.SchemaSearchPath,
				db.config.SchemaMigrationsTable,
			),
		)
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
	db.SetConnMaxLifetime(db.config.ConnMaxLifetime)
	db.SetMaxIdleConns(db.config.MaxIdleConns)
	db.SetMaxOpenConns(db.config.MaxOpenConns)

	return db.Ping()
}

func formatDBQuery(query string) string {
	formattedQuery := strings.Trim(query, "\n")
	formattedQuery = strings.TrimSpace(formattedQuery)

	if strings.Contains(formattedQuery, "\n") {
		formattedQuery = strings.ReplaceAll(formattedQuery, "\n", "\n\t\t\t\t\t     ")
	}

	return loggerDBPrefix + formattedQuery
}
