package database

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/appist/appy/support"
	"github.com/go-pg/pg/v9"
)

// Conn represents a single database connection rather than a pool of database
// connections. Prefer running queries from DB unless there is a specific
// need for a continuous single database connection.
//
// A Conn must call Close to return the connection to the database pool
// and may do so concurrently with a running query.
//
// After a call to Close, all operations on the connection fail.
type Conn = pg.Conn

// Options contains database connection options.
type Options = pg.Options

// ParseDbConfigs retrieves the database config from the environment variables and use it to initialize the database
// connection handler accordingly. For example, to access a database connection handler for `primary` database, i.e.
// appy.Db["primary"], we can have: DB_PRIMARY_ADDR=0.0.0.0:5432 which configures the database connection host/port.
func ParseDbConfigs() map[string]*Options {
	var err error
	dbOptions := map[string]*Options{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		re := regexp.MustCompile("DB_ADDR_(.*)")
		match := re.FindStringSubmatch(val)

		if len(match) > 1 {
			dbNames = append(dbNames, match[1])
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
		if val, ok := os.LookupEnv("DB_DATABASE_" + dbName); ok && val != "" {
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
				support.Logger.Fatal(err)
			}
		}

		retryStatement := false
		if val, ok := os.LookupEnv("DB_RETRY_STATEMENT_" + dbName); ok && val != "" {
			retryStatement, err = strconv.ParseBool(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		poolSize := 10
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			poolSize, err = strconv.Atoi(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		poolTimeout := 31 * time.Second
		if val, ok := os.LookupEnv("DB_POOL_TIMEOUT_" + dbName); ok && val != "" {
			poolTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		minIdleConns := 0
		if val, ok := os.LookupEnv("DB_MIN_IDLE_CONNS_" + dbName); ok && val != "" {
			minIdleConns, err = strconv.Atoi(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		maxConnAge := 0 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_CONN_AGE_" + dbName); ok && val != "" {
			maxConnAge, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		dialTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_DIAL_TIMEOUT_" + dbName); ok && val != "" {
			dialTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		idleCheckFrequency := 1 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_CHECK_FREQUENCY_" + dbName); ok && val != "" {
			idleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		idleTimeout := 5 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_TIMEOUT_" + dbName); ok && val != "" {
			idleTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		readTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			readTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}

			if poolTimeout == 31*time.Second {
				poolTimeout = readTimeout + 1*time.Second
			}
		}

		writeTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_WRITE_TIMEOUT_" + dbName); ok && val != "" {
			writeTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		dbOptions[dbName] = &Options{
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
			OnConnect: func(conn *Conn) error {
				_, err := conn.Exec("SET search_path=?", defaultSchema)
				if err != nil {
					support.Logger.Fatal(err)
				}

				return nil
			},
		}
	}

	return dbOptions
}
