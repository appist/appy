package appy

import (
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
)

// DBManager manages multiple database handles.
type DBManager struct {
	databases map[string]*DB
	errors    []error
	logger    *Logger
}

// NewDBManager initializes DbManager instance.
func NewDBManager(logger *Logger, support Supporter) *DBManager {
	dbManager := &DBManager{
		databases: map[string]*DB{},
		logger:    logger,
	}
	dbConfig, errs := parseDBConfig(support)
	if errs != nil {
		dbManager.errors = errs
	}

	for name, config := range dbConfig {
		dbManager.databases[name] = NewDB(config, logger, support)
	}

	return dbManager
}

// DB returns the database handle with the specified name.
func (m *DBManager) DB(name string) *DB {
	if db, ok := m.databases[name]; ok {
		return db
	}

	return nil
}

// Errors returns the DB manager errors.
func (m *DBManager) Errors() []error {
	return m.errors
}

// Info returns the DB manager info.
func (m *DBManager) Info() string {
	var dbNames []string
	for name := range m.databases {
		dbNames = append(dbNames, name)
	}

	databases := "none"
	if len(dbNames) > 0 {
		databases = strings.Join(dbNames, ", ")
	}

	return fmt.Sprintf("* DBs: %s", databases)
}

func parseDBConfig(support Supporter) (map[string]*DBConfig, []error) {
	var (
		err  error
		errs []error
	)
	dbConfig := map[string]*DBConfig{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		addrRegex := regexp.MustCompile("DB_ADDR_(.*)")
		addrMatch := addrRegex.FindStringSubmatch(val)

		if len(addrMatch) > 1 {
			splits := strings.Split(addrMatch[1], "=")
			dbNames = append(dbNames, splits[0])
		}

		urlRegex := regexp.MustCompile("DB_URL_(.*)")
		urlMatch := urlRegex.FindStringSubmatch(val)

		if len(urlMatch) > 1 {
			splits := strings.Split(urlMatch[1], "=")

			if support.ArrayContains(dbNames, splits[0]) {
				errs = append(errs, fmt.Errorf("please only specify either 'DB_ADDR_%s' or 'DB_URL_%s'", splits[0], splits[0]))
				continue
			}

			dbNames = append(dbNames, splits[0])
		}
	}

	for _, dbName := range dbNames {
		config := &DBConfig{}

		config.SchemaSearchPath = "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_SEARCH_PATH_" + dbName); ok && val != "" {
			config.SchemaSearchPath = val
		}

		config.Network = "tcp"
		if val, ok := os.LookupEnv("DB_NETWORK_" + dbName); ok && val != "" {
			config.Network = val
		}

		config.Addr = "0.0.0.0:5432"
		if val, ok := os.LookupEnv("DB_ADDR_" + dbName); ok && val != "" {
			config.Addr = val
		}

		config.User = "postgres"
		if val, ok := os.LookupEnv("DB_USER_" + dbName); ok && val != "" {
			config.User = val
		}

		config.Password = "postgres"
		if val, ok := os.LookupEnv("DB_PASSWORD_" + dbName); ok && val != "" {
			config.Password = val
		}

		config.Database = "postgres"
		if val, ok := os.LookupEnv("DB_DATABASE_" + dbName); ok && val != "" {
			config.Database = val
		}

		config.ApplicationName = "appy"
		if val, ok := os.LookupEnv("DB_APPLICATION_NAME_" + dbName); ok && val != "" {
			config.ApplicationName = val
		}

		config.Replica = false
		if val, ok := os.LookupEnv("DB_REPLICA_" + dbName); ok && val != "" {
			config.Replica, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MaxRetries = 0
		if val, ok := os.LookupEnv("DB_MAX_RETRIES_" + dbName); ok && val != "" {
			config.MaxRetries, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.RetryStatementTimeout = false
		if val, ok := os.LookupEnv("DB_RETRY_STATEMENT_" + dbName); ok && val != "" {
			config.RetryStatementTimeout, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MinRetryBackoff = 250 * time.Millisecond
		if val, ok := os.LookupEnv("DB_MIN_RETRY_BACKOFF_" + dbName); ok && val != "" {
			config.MinRetryBackoff, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MaxRetryBackoff = 4 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_RETRY_BACKOFF_" + dbName); ok && val != "" {
			config.MaxRetryBackoff, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.PoolSize = 10
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			config.PoolSize, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.PoolTimeout = 10 * time.Second
		if val, ok := os.LookupEnv("DB_POOL_TIMEOUT_" + dbName); ok && val != "" {
			config.PoolTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MinIdleConns = 0
		if val, ok := os.LookupEnv("DB_MIN_IDLE_CONNS_" + dbName); ok && val != "" {
			config.MinIdleConns, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MaxConnAge = 0 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_CONN_AGE_" + dbName); ok && val != "" {
			config.MaxConnAge, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.DialTimeout = 5 * time.Second
		if val, ok := os.LookupEnv("DB_DIAL_TIMEOUT_" + dbName); ok && val != "" {
			config.DialTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.IdleCheckFrequency = 1 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_CHECK_FREQUENCY_" + dbName); ok && val != "" {
			config.IdleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.IdleTimeout = 5 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_TIMEOUT_" + dbName); ok && val != "" {
			config.IdleTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.ReadTimeout = 10 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			config.ReadTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.WriteTimeout = 10 * time.Second
		if val, ok := os.LookupEnv("DB_WRITE_TIMEOUT_" + dbName); ok && val != "" {
			config.WriteTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.SchemaMigrationsTable = "schema_migrations"
		if val, ok := os.LookupEnv("DB_SCHEMA_MIGRATIONS_TABLE_" + dbName); ok && val != "" {
			config.SchemaMigrationsTable = val
		}

		config.TLSConfig = nil
		if val, ok := os.LookupEnv("DB_SSLMODE_" + dbName); ok && val != "" && val != "disable" {
			switch val {
			case "verify-ca", "verify-full":
				config.TLSConfig = &tls.Config{}
			case "allow", "prefer", "require":
				config.TLSConfig = &tls.Config{InsecureSkipVerify: true} //nolint
			default:
				errs = append(errs, fmt.Errorf("sslmode '%v' is not supported", val))
			}
		}

		config.OnConnect = func(conn *DBConn) error {
			_, err := conn.Exec("SET search_path=? ?", DBSafeQuery(config.SchemaSearchPath), DBSafeQuery(dbQueryComment))
			if err != nil {
				return err
			}

			return nil
		}

		if os.Getenv("DB_URL_"+dbName) != "" {
			opts, err := pg.ParseURL(os.Getenv("DB_URL_" + dbName))
			if err != nil {
				errs = append(errs, err)
			} else {
				config.Addr = opts.Addr
				config.User = opts.User
				config.Password = opts.Password
				config.Database = opts.Database
				config.DialTimeout = opts.DialTimeout
				config.TLSConfig = opts.TLSConfig
			}
		}

		dbConfig[support.ToCamelCase(dbName)] = config
	}

	return dbConfig, errs
}
