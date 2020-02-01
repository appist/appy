package appy

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DBManager manages multiple database handles.
type DBManager struct {
	databases map[string]*DB
	errors    []error
	logger    *Logger
}

// NewDbManager initializes DbManager instance.
func NewDbManager(logger *Logger, support Supporter) *DBManager {
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
		re := regexp.MustCompile("DB_ADDR_(.*)")
		match := re.FindStringSubmatch(val)

		if len(match) > 1 {
			splits := strings.Split(match[1], "=")
			dbNames = append(dbNames, splits[0])
		}
	}

	for _, dbName := range dbNames {
		config := &DBConfig{}

		config.SchemaSearchPath = "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_SEARCH_PATH_" + dbName); ok && val != "" {
			config.SchemaSearchPath = val
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

		config.PoolSize = 25
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			config.PoolSize, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.PoolTimeout = 30 * time.Second
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

		config.DialTimeout = 30 * time.Second
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

		config.ReadTimeout = 30 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			config.ReadTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.WriteTimeout = 30 * time.Second
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

		dbConfig[support.ToCamelCase(dbName)] = config
	}

	return dbConfig, errs
}
