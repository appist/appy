package database

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/appist/appy/support"
	"github.com/go-pg/pg/v9"
)

type Conn = pg.Conn
type Options = pg.Options

// ParseDbConfigs retrieves the database config from the environment variables and use it to initialize the database
// connection handler accordingly. For example, to access a database connection handler for `primary` database, i.e.
// appy.Db["primary"], we can have: DB_PRIMARY_ADDR=0.0.0.0:5432 which configures the database connection host/port.
func ParseDbConfigs() map[string]*Options {
	var err error
	dbOptions := map[string]*Options{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		re := regexp.MustCompile("DB_(.*)_ADDR")
		match := re.FindStringSubmatch(val)

		if len(match) > 1 {
			dbNames = append(dbNames, match[1])
		}
	}

	for _, dbName := range dbNames {
		defaultSchema := "public"
		if val, ok := os.LookupEnv("DB_" + dbName + "_DEFAULT_SCHEMA"); ok && val != "" {
			defaultSchema = val
		}

		addr := "0.0.0.0:5432"
		if val, ok := os.LookupEnv("DB_" + dbName + "_ADDR"); ok && val != "" {
			addr = val
		}

		user := "postgres"
		if val, ok := os.LookupEnv("DB_" + dbName + "_USER"); ok && val != "" {
			user = val
		}

		password := "postgres"
		if val, ok := os.LookupEnv("DB_" + dbName + "_PASSWORD"); ok && val != "" {
			password = val
		}

		database := "appy"
		if val, ok := os.LookupEnv("DB_" + dbName + "_DATABASE"); ok && val != "" {
			database = val
		}

		appName := "appy"
		if val, ok := os.LookupEnv("DB_" + dbName + "_APP_NAME"); ok && val != "" {
			appName = val
		}

		maxRetries := 0
		if val, ok := os.LookupEnv("DB_" + dbName + "_MAX_RETRIES"); ok && val != "" {
			maxRetries, err = strconv.Atoi(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		retryStatement := false
		if val, ok := os.LookupEnv("DB_" + dbName + "_RETRY_STATEMENT"); ok && val != "" {
			retryStatement, err = strconv.ParseBool(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		poolSize := 10
		if val, ok := os.LookupEnv("DB_" + dbName + "_POOL_SIZE"); ok && val != "" {
			poolSize, err = strconv.Atoi(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		poolTimeout := 31 * time.Second
		if val, ok := os.LookupEnv("DB_" + dbName + "_POOL_TIMEOUT"); ok && val != "" {
			poolTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		minIdleConns := 0
		if val, ok := os.LookupEnv("DB_" + dbName + "_MIN_IDLE_CONNS"); ok && val != "" {
			minIdleConns, err = strconv.Atoi(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		maxConnAge := 0 * time.Second
		if val, ok := os.LookupEnv("DB_" + dbName + "_MAX_CONN_AGE"); ok && val != "" {
			maxConnAge, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		dialTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_" + dbName + "_DIAL_TIMEOUT"); ok && val != "" {
			dialTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		idleCheckFrequency := 1 * time.Minute
		if val, ok := os.LookupEnv("DB_" + dbName + "_IDLE_CHECK_FREQUENCY"); ok && val != "" {
			idleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		idleTimeout := 5 * time.Minute
		if val, ok := os.LookupEnv("DB_" + dbName + "_IDLE_TIMEOUT"); ok && val != "" {
			idleTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}
		}

		readTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_" + dbName + "_READ_TIMEOUT"); ok && val != "" {
			readTimeout, err = time.ParseDuration(val)
			if err != nil {
				support.Logger.Fatal(err)
			}

			if poolTimeout == 31*time.Second {
				poolTimeout = readTimeout + 1*time.Second
			}
		}

		writeTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_" + dbName + "_WRITE_TIMEOUT"); ok && val != "" {
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
