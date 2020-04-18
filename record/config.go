package record

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/appist/appy/support"
)

// Config contains database connection configuration.
type Config struct {
	// Adapter indicates the database adapter to use. The value is parsed from
	// "DB_URI_<DB_NAME>".
	Adapter string

	// ConnMaxLifetime indicates the maximum amount of time a connection may be
	// reused. Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are reused forever.
	ConnMaxLifetime time.Duration

	// Database indicates the database schema to connect. The value is parsed
	// from "DB_URI_<DB_NAME>".
	Database string

	// Host indicates the host to use for connecting to the database. The value
	// is parsed from "DB_URI_<DB_NAME>".
	Host string

	// MaxIdleConns indicates the maximum number of connections in the idle
	// connection pool. By default, it is 25. Otherwise, the value is parsed
	// from "DB_MAX_IDLE_CONNS_<DB_NAME>".
	//
	// Note: MaxIdleConns will automatically be updated to use the same value
	// as MaxOpenConns if  MaxIdleConns is greater than MaxOpenConns.
	MaxIdleConns int

	// MaxOpenConns indicates the maximum number of open connections to the
	// database. By default, it is 25. Otherwise, the value is parsed
	// from "DB_MAX_OPEN_CONNS_<DB_NAME>".
	MaxOpenConns int

	// Password indicates the password to use for connecting to the database.
	// The value is parsed from "DB_URI_<DB_NAME>".
	Password string

	// Port indicates the port to use for connecting to the database. The value
	// is parsed from "DB_URI_<DB_NAME>".
	Port string

	// Replica indicates if the database is a read replica. By default, it is
	// false. Otherwise, the value is parsed from "DB_REPLICA_<DB_NAME>".
	Replica bool

	// SchemaSearchPath indicates the schema search path which is only used with
	// "postgres" adapter.
	//
	// By default, it is "public". Otherwise, the value is parsed from
	// "DB_SCHEMA_SEARCH_PATH_<DB_NAME>".
	SchemaSearchPath string

	// SchemaMigrationsTable indicates the table name for storing the schema
	// migration versions.
	//
	// By default, it is "schema_migrations". Otherwise, the value is parsed from
	// "DB_SCHEMA_MIGRATIONS_TABLE_<DB_NAME>".
	SchemaMigrationsTable string

	// URI indicates the database connection string to connect.
	//
	// URI connection string documentation:
	//   - mysql: https://dev.mysql.com/doc/refman/8.0/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri
	//   - postgres: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	URI string

	// Username indicates the username to use for connecting to the database. The
	// value is parsed from "DB_URI_<DB_NAME>".
	Username string
}

func (c *Config) parseDBInfoFromURI() (err error) {
	uri := c.URI
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	newURI := uri
	scheme := u.Scheme
	switch u.Scheme {
	case "mysql":
		newURI = strings.ReplaceAll(uri, u.Host, "tcp("+u.Host+")")
		newURI = strings.ReplaceAll(newURI, "mysql://", "")
	case "postgres", "postgresql":
		scheme = "postgres"
	}

	c.Adapter = scheme
	c.Database = strings.Trim(u.Path, "/")
	c.Host = u.Hostname()
	c.Password, _ = u.User.Password()
	c.Port = u.Port()
	c.URI = newURI
	c.Username = u.User.Username()
	return nil
}

func parseDBConfig() (map[string]*Config, []error) {
	var errs []error
	dbConfig := map[string]*Config{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		uriRegex := regexp.MustCompile("DB_URI_(.*)")
		uriMatches := uriRegex.FindStringSubmatch(val)

		if len(uriMatches) > 1 {
			splits := strings.Split(uriMatches[1], "=")
			dbNames = append(dbNames, splits[0])
		}
	}

	for _, dbName := range dbNames {
		var err error
		config := &Config{}

		config.ConnMaxLifetime = 5 * time.Minute
		if val, ok := os.LookupEnv("DB_CONN_MAX_LIFETIME_" + dbName); ok && val != "" {
			config.ConnMaxLifetime, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MaxIdleConns = 25
		if val, ok := os.LookupEnv("DB_MAX_IDLE_CONNS_" + dbName); ok && val != "" {
			config.MaxIdleConns, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.MaxOpenConns = 25
		if val, ok := os.LookupEnv("DB_MAX_OPEN_CONNS_" + dbName); ok && val != "" {
			config.MaxOpenConns, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		if config.MaxIdleConns > config.MaxOpenConns {
			config.MaxIdleConns = config.MaxOpenConns
		}

		config.Replica = false
		if val, ok := os.LookupEnv("DB_REPLICA_" + dbName); ok && val != "" {
			config.Replica, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		config.SchemaMigrationsTable = "schema_migrations"
		if val, ok := os.LookupEnv("DB_SCHEMA_MIGRATIONS_TABLE_" + dbName); ok && val != "" {
			config.SchemaMigrationsTable = val
		}

		config.SchemaSearchPath = "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_SEARCH_PATH_" + dbName); ok && val != "" {
			config.SchemaSearchPath = val
		}

		if val, ok := os.LookupEnv("DB_URI_" + dbName); ok && val != "" {
			config.URI = val
		}

		if config.URI != "" {
			err = config.parseDBInfoFromURI()
			if err != nil {
				errs = append(errs, err)
				continue
			}

			if !support.ArrayContains(support.SupportedDBAdapters, config.Adapter) {
				errs = append(errs, fmt.Errorf("adapter '%s' for database '%s' is not supported", config.Adapter, support.ToCamelCase(dbName)))
				continue
			}
		}

		dbConfig[support.ToCamelCase(dbName)] = config
	}

	return dbConfig, errs
}
