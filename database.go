package appy

import (
	"os"
	"sync"

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

	DbManagerer interface {
	}

	// Db provides the functionality to communicate with the database.
	Db struct {
		config     DbConfig
		handle     *DbHandle
		logger     *Logger
		migrations []*DbMigration
		mu         sync.Mutex
	}

	// DbManager manages multiple databases.
	DbManager struct {
		dbs    map[string]*Db
		errors []error
		mu     sync.Mutex
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

// NewDbManager initializes DbManager instance.
func NewDbManager(logger *Logger, support *Support) *DbManager {
	dbManager := &DbManager{
		dbs: map[string]*Db{},
	}
	dbConfig, errs := parseDbConfig(support)
	if errs != nil {
		dbManager.errors = errs
	}

	for name, val := range dbConfig {
		dbManager.dbs[name] = &Db{
			config: val,
			logger: logger,
		}
	}

	return dbManager
}

// Connect establishes connections to all the databases.
func (m *DbManager) Connect(sameDb bool) error {
	for _, db := range m.dbs {
		err := db.Connect(sameDb)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes connections to all the databases.
func (m *DbManager) Close() error {
	for _, db := range m.dbs {
		err := db.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Db returns the Db instance with the specified name.
func (m *DbManager) Db(name string) *Db {
	if db, ok := m.dbs[name]; ok {
		return db
	}

	return nil
}

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

// CreateDb creates the database.
func (db *Db) CreateDb() []error {
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

// DropDb creates the database.
func (db *Db) DropDb() []error {
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

// CheckDbManager is used to check if DB manager contains any error during initialization.
func CheckDbManager(config *Config, dbManager *DbManager, logger *Logger) {
	if dbManager != nil && dbManager.errors != nil {
		for _, err := range dbManager.errors {
			logger.Info(err.Error())
		}

		os.Exit(-1)
	}
}
