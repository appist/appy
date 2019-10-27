package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/appist/appy/support"
	"github.com/go-pg/pg/v9"
)

// AppDbConn represents a single database connection rather than a pool of database
// connections. Prefer running queries from DB unless there is a specific
// need for a continuous single database connection.
//
// A Conn must call Close to return the connection to the database pool
// and may do so concurrently with a running query.
//
// After a call to Close, all operations on the connection fail.
type AppDbConn = pg.Conn

// AppDbConfig keeps database connection options.
type AppDbConfig struct {
	pg.Options
	Replica               bool
	Schema                string
	SchemaMigrationsTable string
}

// AppDbHandler is a database handle representing a pool of zero or more underlying connections. It's safe
// for concurrent use by multiple goroutines.
type AppDbHandler = pg.DB

// AppDbHandlerTx is an in-progress database transaction. It is safe for concurrent use by multiple goroutines.
type AppDbHandlerTx = pg.Tx

// AppDb keeps database connection with its configuration.
type AppDb struct {
	Config     AppDbConfig
	Handler    *AppDbHandler
	Logger     *AppLogger
	Migrations []*AppDbMigration
	mu         sync.Mutex
}

// AppDbMigration keeps database migration options.
type AppDbMigration struct {
	File    string
	Version string
	Down    func(*AppDbHandler) error
	DownTx  func(*AppDbHandlerTx) error
	Up      func(*AppDbHandler) error
	UpTx    func(*AppDbHandlerTx) error
}

// AppDbSchemaMigration is the model that maps to schema_migrations table.
type AppDbSchemaMigration struct {
	Version string
}

// AppDbQueryEvent keeps the query event information.
type AppDbQueryEvent = pg.QueryEvent

// AppDbTx is an in-progress database transaction. It is safe for concurrent use by multiple goroutines.
type AppDbTx = pg.Tx

var (
	// Array accepts a slice and returns a wrapper for working with PostgreSQL array data type.
	Array = pg.Array

	// SafeQuery replaces any placeholders found in the query.
	SafeQuery = pg.SafeQuery

	// Scan returns ColumnScanner that copies the columns in the row into the values.
	Scan = pg.Scan
)

func parseDbConfig() (map[string]AppDbConfig, error) {
	var err error
	dbConfig := map[string]AppDbConfig{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		re := regexp.MustCompile("DB_ADDR_(.*)")
		match := re.FindStringSubmatch(val)

		if len(match) > 1 {
			splits := strings.Split(match[1], "=")
			dbNames = append(dbNames, splits[0])
		}
	}

	for _, dbName := range dbNames {
		schema := "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_" + dbName); ok && val != "" {
			schema = val
		}

		addr := "0.0.0.0:5432"
		if val, ok := os.LookupEnv("DB_ADDR_" + dbName); ok && val != "" {
			addr = val
		}

		user := "postgres"
		if val, ok := os.LookupEnv("DB_USER_" + dbName); ok && val != "" {
			user = val
		}

		password := "postgres"
		if val, ok := os.LookupEnv("DB_PASSWORD_" + dbName); ok && val != "" {
			password = val
		}

		database := "appy"
		if val, ok := os.LookupEnv("DB_NAME_" + dbName); ok && val != "" {
			database = val
		}

		appName := "appy"
		if val, ok := os.LookupEnv("DB_APP_NAME_" + dbName); ok && val != "" {
			appName = val
		}

		replica := false
		if val, ok := os.LookupEnv("DB_REPLICA_" + dbName); ok && val != "" {
			replica, err = strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}
		}

		maxRetries := 0
		if val, ok := os.LookupEnv("DB_MAX_RETRIES_" + dbName); ok && val != "" {
			maxRetries, err = strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
		}

		retryStatement := false
		if val, ok := os.LookupEnv("DB_RETRY_STATEMENT_" + dbName); ok && val != "" {
			retryStatement, err = strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}
		}

		poolSize := 10
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			poolSize, err = strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
		}

		poolTimeout := 31 * time.Second
		if val, ok := os.LookupEnv("DB_POOL_TIMEOUT_" + dbName); ok && val != "" {
			poolTimeout, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		minIdleConns := 0
		if val, ok := os.LookupEnv("DB_MIN_IDLE_CONNS_" + dbName); ok && val != "" {
			minIdleConns, err = strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
		}

		maxConnAge := 0 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_CONN_AGE_" + dbName); ok && val != "" {
			maxConnAge, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		dialTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_DIAL_TIMEOUT_" + dbName); ok && val != "" {
			dialTimeout, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		idleCheckFrequency := 1 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_CHECK_FREQUENCY_" + dbName); ok && val != "" {
			idleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		idleTimeout := 5 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_TIMEOUT_" + dbName); ok && val != "" {
			idleTimeout, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		readTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			readTimeout, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}

			if poolTimeout == 31*time.Second {
				poolTimeout = readTimeout + 1*time.Second
			}
		}

		writeTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_WRITE_TIMEOUT_" + dbName); ok && val != "" {
			writeTimeout, err = time.ParseDuration(val)
			if err != nil {
				return nil, err
			}
		}

		schemaMigrationsTable := "schema_migrations"
		if val, ok := os.LookupEnv("DB_SCHEMA_MIGRATIONS_TABLE_" + dbName); ok && val != "" {
			schemaMigrationsTable = val
		}

		config := AppDbConfig{}
		config.ApplicationName = appName
		config.Addr = addr
		config.User = user
		config.Password = password
		config.Database = database
		config.MaxRetries = maxRetries
		config.PoolSize = poolSize
		config.PoolTimeout = poolTimeout
		config.DialTimeout = dialTimeout
		config.IdleCheckFrequency = idleCheckFrequency
		config.IdleTimeout = idleTimeout
		config.ReadTimeout = readTimeout
		config.WriteTimeout = writeTimeout
		config.RetryStatementTimeout = retryStatement
		config.MinIdleConns = minIdleConns
		config.MaxConnAge = maxConnAge
		config.Replica = replica
		config.Schema = schema
		config.SchemaMigrationsTable = schemaMigrationsTable
		config.OnConnect = func(conn *AppDbConn) error {
			_, err := conn.Exec("SET search_path=?", schema)
			if err != nil {
				return err
			}

			return nil
		}

		dbConfig[strings.ToLower(dbName)] = config
	}

	return dbConfig, nil
}

func newDb(config AppDbConfig, logger *AppLogger) (*AppDb, error) {
	return &AppDb{
		Config: config,
		Logger: logger,
	}, nil
}

// DbConnect establishes connections to all the databases.
func DbConnect(dbMap map[string]*AppDb, sameDb bool) error {
	for _, db := range dbMap {
		err := db.Connect(sameDb)
		if err != nil {
			return err
		}
	}

	return nil
}

// DbClose closes connections to all the databases.
func DbClose(dbMap map[string]*AppDb) error {
	for _, db := range dbMap {
		db.Close()
	}

	return nil
}

// Connect connects to a database using provided options and assign the database Handler which is safe for concurrent
// use by multiple goroutines and maintains its own connection pool.
func (db *AppDb) Connect(sameDb bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	opts := db.Config.Options
	if !sameDb {
		opts.Database = "postgres"
	}

	db.Handler = pg.Connect(&opts)
	db.Handler.AddQueryHook(db.Logger)
	_, err := db.Handler.Exec("SELECT 1")
	return err
}

// Close closes the database connection and release any open resources. It is rare to Close a DB, as the DB handle is
// meant to be long-lived and shared between many goroutines.
func (db *AppDb) Close() {
	db.Handler.Close()
}

// Migrate runs migrations for the current environment that have not run yet.
func (db *AppDb) Migrate() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.ensureSchemaMigrationsTable()
	if err != nil {
		return err
	}

	tx, err := db.Handler.Begin()
	defer tx.Close()
	if err != nil {
		return err
	}

	migratedVersions, err := db.migratedVersions(tx)
	if err != nil {
		return err
	}

	for _, m := range db.Migrations {
		if !support.ArrayContains(migratedVersions, m.Version) {
			if m.UpTx != nil {
				err = m.UpTx(tx)
				if err != nil {
					return err
				}

				err = db.addSchemaMigrations(nil, tx, m)
				if err != nil {
					return err
				}

				continue
			}

			err = m.Up(db.Handler)
			if err != nil {
				return err
			}

			err = db.addSchemaMigrations(db.Handler, nil, m)
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

func (db *AppDb) migratedVersions(tx *AppDbTx) ([]string, error) {
	var schemaMigrations []AppDbSchemaMigration
	_, err := tx.Query(&schemaMigrations, `SELECT version FROM ?.? ORDER BY version ASC`, SafeQuery(db.Config.Schema), SafeQuery(db.Config.SchemaMigrationsTable))
	if err != nil {
		return nil, err
	}

	migratedVersions := []string{}
	for _, m := range schemaMigrations {
		migratedVersions = append(migratedVersions, m.Version)
	}

	return migratedVersions, nil
}

func (db *AppDb) ensureSchemaMigrationsTable() error {
	count, err := db.Handler.
		Model().
		Table("pg_tables").
		Where("schemaname = '?'", SafeQuery(db.Config.Schema)).
		Where("tablename = '?'", SafeQuery(db.Config.SchemaMigrationsTable)).
		Count()

	if err != nil {
		return err
	}

	if count < 1 {
		_, err = db.Handler.Exec(`CREATE SCHEMA IF NOT EXISTS ?`, SafeQuery(db.Config.Schema))
		if err != nil {
			return err
		}

		_, err = db.Handler.Exec(`CREATE TABLE ? (version VARCHAR PRIMARY KEY)`, SafeQuery(db.Config.SchemaMigrationsTable))
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterMigration registers the up/down migrations that won't be executed in transaction.
func (db *AppDb) RegisterMigration(up func(*AppDbHandler) error, down func(*AppDbHandler) error) {
	err := db.registerMigration(up, down, nil, nil)
	if err != nil {
		db.Logger.Fatal(err)
	}
}

// RegisterMigrationTx registers the up/down migrations that will be executed in transaction.
func (db *AppDb) RegisterMigrationTx(upTx func(*AppDbHandlerTx) error, downTx func(*AppDbHandlerTx) error) {
	err := db.registerMigration(nil, nil, upTx, downTx)
	if err != nil {
		db.Logger.Fatal(err)
	}
}

func (db *AppDb) registerMigration(up func(*AppDbHandler) error, down func(*AppDbHandler) error, upTx func(*AppDbHandlerTx) error, downTx func(*AppDbHandlerTx) error) error {
	file := migrationFile()
	version, err := migrationVersion(file)
	if err != nil {
		return err
	}

	db.addMigration(&AppDbMigration{
		File:    file,
		Version: version,
		Down:    down,
		DownTx:  downTx,
		Up:      up,
		UpTx:    upTx,
	})

	return nil
}

func (db *AppDb) addMigration(newMigration *AppDbMigration) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Migrations = append(db.Migrations, newMigration)
}

func (db *AppDb) addSchemaMigrations(handler *AppDbHandler, handlerTx *AppDbHandlerTx, newMigration *AppDbMigration) error {
	query := `INSERT INTO ?.? (version) VALUES (?)`
	if handler != nil {
		_, err := handler.Exec(
			query,
			SafeQuery(db.Config.Schema),
			SafeQuery(db.Config.SchemaMigrationsTable),
			SafeQuery(newMigration.Version),
		)
		return err
	}

	_, err := handlerTx.Exec(
		query,
		SafeQuery(db.Config.Schema),
		SafeQuery(db.Config.SchemaMigrationsTable),
		SafeQuery(newMigration.Version),
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
