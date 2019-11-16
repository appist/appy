package orm

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	appysupport "github.com/appist/appy/internal/support"
)

type (
	// DbManagerer implements all the DbManager's methods.
	DbManagerer interface {
	}

	// DbManager manages multiple databases.
	DbManager struct {
		dbs    map[string]*Db
		errors []error
		logger *appysupport.Logger
		mu     sync.Mutex
	}
)

// NewDbManager initializes DbManager instance.
func NewDbManager(logger *appysupport.Logger) *DbManager {
	dbManager := &DbManager{
		dbs:    map[string]*Db{},
		logger: logger,
	}
	dbConfig, errs := parseDbConfig()
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

// ConnectAll establishes connections to all the databases.
func (m *DbManager) ConnectAll(sameDb bool) error {
	for _, db := range m.dbs {
		err := db.Connect(sameDb)
		if err != nil {
			return err
		}
	}

	return nil
}

// CloseAll closes connections to all the databases.
func (m *DbManager) CloseAll() error {
	for _, db := range m.dbs {
		err := db.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Db returns the database instance with the specified name.
func (m *DbManager) Db(name string) *Db {
	if db, ok := m.dbs[name]; ok {
		return db
	}

	return nil
}

// Dbs returns all the database instances.
func (m *DbManager) Dbs() map[string]*Db {
	return m.dbs
}

// DbHandle returns the handle for the database instance with the specified name.
func (m *DbManager) DbHandle(name string) *DbHandle {
	if db, ok := m.dbs[name]; ok {
		return db.handle
	}

	return nil
}

// Errors returns all the DB manager setup errors.
func (m *DbManager) Errors() []error {
	return m.errors
}

// PrintInfo prints the database manager info.
func (m *DbManager) PrintInfo() {
	var dbNames []string
	for name := range m.dbs {
		dbNames = append(dbNames, name)
	}

	dbs := "none"
	if len(dbNames) > 0 {
		dbs = strings.Join(dbNames, ", ")
	}

	m.logger.Infof("* Available DBs: %s", dbs)
}

// IsDbManagerErrored is used to check if DB manager contains any error during initialization.
func IsDbManagerErrored(config *appysupport.Config, dbManager *DbManager, logger *appysupport.Logger) bool {
	if dbManager != nil && dbManager.errors != nil {
		for _, err := range dbManager.errors {
			logger.Info(err.Error())
		}

		return true
	}

	return false
}

func parseDbConfig() (map[string]DbConfig, []error) {
	var (
		err  error
		errs []error
	)
	dbConfig := map[string]DbConfig{}
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
		schemaSearchPath := "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_SEARCH_PATH_" + dbName); ok && val != "" {
			schemaSearchPath = val
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
				errs = append(errs, err)
			}
		}

		maxRetries := 0
		if val, ok := os.LookupEnv("DB_MAX_RETRIES_" + dbName); ok && val != "" {
			maxRetries, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		retryStatement := false
		if val, ok := os.LookupEnv("DB_RETRY_STATEMENT_" + dbName); ok && val != "" {
			retryStatement, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		poolSize := 10
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			poolSize, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		poolTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_POOL_TIMEOUT_" + dbName); ok && val != "" {
			poolTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		minIdleConns := 0
		if val, ok := os.LookupEnv("DB_MIN_IDLE_CONNS_" + dbName); ok && val != "" {
			minIdleConns, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		maxConnAge := 0 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_CONN_AGE_" + dbName); ok && val != "" {
			maxConnAge, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		dialTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_DIAL_TIMEOUT_" + dbName); ok && val != "" {
			dialTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		idleCheckFrequency := 1 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_CHECK_FREQUENCY_" + dbName); ok && val != "" {
			idleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		idleTimeout := 5 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_TIMEOUT_" + dbName); ok && val != "" {
			idleTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		readTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			readTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		writeTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_WRITE_TIMEOUT_" + dbName); ok && val != "" {
			writeTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		schemaMigrationsTable := "schema_migrations"
		if val, ok := os.LookupEnv("DB_SCHEMA_MIGRATIONS_TABLE_" + dbName); ok && val != "" {
			schemaMigrationsTable = val
		}

		config := DbConfig{}
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
		config.SchemaSearchPath = schemaSearchPath
		config.SchemaMigrationsTable = schemaMigrationsTable
		config.OnConnect = func(conn *DbConn) error {
			_, err := conn.Exec("SET search_path=? ?", DbSafeQuery(schemaSearchPath), DbSafeQuery(appysupport.DbQueryComment))
			if err != nil {
				return err
			}

			return nil
		}

		dbConfig[appysupport.ToCamelCase(dbName)] = config
	}

	return dbConfig, errs
}
