package core

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
type AppDbConfig = pg.Options

// AppDbHandler is a database handle representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type AppDbHandler = pg.DB

// AppDb keeps database connection with its configuration.
type AppDb struct {
	Config  AppDbConfig
	Handler *AppDbHandler
}

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
		defaultSchema := "public"
		if val, ok := os.LookupEnv("DB_DEFAULT_SCHEMA_" + dbName); ok && val != "" {
			defaultSchema = val
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

		dbConfig[dbName] = AppDbConfig{
			ApplicationName:       appName,
			Addr:                  addr,
			User:                  user,
			Password:              password,
			Database:              database,
			MaxRetries:            maxRetries,
			PoolSize:              poolSize,
			PoolTimeout:           poolTimeout,
			DialTimeout:           dialTimeout,
			IdleCheckFrequency:    idleCheckFrequency,
			IdleTimeout:           idleTimeout,
			ReadTimeout:           readTimeout,
			WriteTimeout:          writeTimeout,
			RetryStatementTimeout: retryStatement,
			MinIdleConns:          minIdleConns,
			MaxConnAge:            maxConnAge,
			OnConnect: func(conn *AppDbConn) error {
				_, err := conn.Exec("SET search_path=?", defaultSchema)
				if err != nil {
					return err
				}

				return nil
			},
		}
	}

	return dbConfig, nil
}

func newDb(config AppDbConfig) (*AppDb, error) {
	return &AppDb{
		Config: config,
	}, nil
}

// Connect connects to a database using provided options and assign the database Handler which is safe for concurrent
// use by multiple goroutines and maintains its own connection pool.
func (db *AppDb) Connect() error {
	db.Handler = pg.Connect(&db.Config)
	_, err := db.Handler.Exec("SELECT 1")
	return err
}

// Close closes the database connection and release any open resources. It is rare to Close a DB, as the DB handle is
// meant to be long-lived and shared between many goroutines.
func (db *AppDb) Close() {
	db.Handler.Close()
}
