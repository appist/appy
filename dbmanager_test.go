package appy

import (
	"os"
	"testing"
	"time"
)

type DBManagerSuite struct {
	TestSuite
	logger  *Logger
	support Supporter
}

func (s *DBManagerSuite) SetupTest() {
	s.logger, _, _ = NewFakeLogger()
	s.support = &Support{}
}

func (s *DBManagerSuite) TearDownTest() {
}

func (s *DBManagerSuite) TestNewDBManagerWithDefaultConfig() {
	addr := os.Getenv("DB_ADDR_PRIMARY")
	if addr != "" {
		os.Unsetenv("DB_ADDR_PRIMARY")
	}

	os.Setenv("DB_ADDR_MAIN_APP", "0.0.0.0:5432")
	defer func() {
		os.Setenv("DB_ADDR_PRIMARY", addr)
		os.Unsetenv("DB_ADDR_MAIN_APP")
	}()

	dbManager := NewDBManager(s.logger, s.support)
	s.Nil(dbManager.DB("primary"))
	s.NotNil(dbManager.DB("mainApp"))
	s.Equal(0, len(dbManager.Errors()))
	s.Equal("* DBs: mainApp", dbManager.Info())

	config := dbManager.DB("mainApp").Config()
	s.Equal("public", config.SchemaSearchPath)
	s.Equal("tcp", config.Network)
	s.Equal("0.0.0.0:5432", config.Addr)
	s.Equal("postgres", config.User)
	s.Equal("postgres", config.Password)
	s.Equal("postgres", config.Database)
	s.Equal("appy", config.ApplicationName)
	s.Equal(false, config.Replica)
	s.Equal(0, config.MaxRetries)
	s.Equal(false, config.RetryStatementTimeout)
	s.Equal(250*time.Millisecond, config.MinRetryBackoff)
	s.Equal(4*time.Second, config.MaxRetryBackoff)
	s.Equal(10, config.PoolSize)
	s.Equal(10*time.Second, config.PoolTimeout)
	s.Equal(0, config.MinIdleConns)
	s.Equal(0*time.Second, config.MaxConnAge)
	s.Equal(5*time.Second, config.DialTimeout)
	s.Equal(1*time.Minute, config.IdleCheckFrequency)
	s.Equal(5*time.Minute, config.IdleTimeout)
	s.Equal(10*time.Second, config.ReadTimeout)
	s.Equal(10*time.Second, config.WriteTimeout)
	s.Equal("schema_migrations", config.SchemaMigrationsTable)
}

func (s *DBManagerSuite) TestNewDBManagerWithNoConfig() {
	addr := os.Getenv("DB_ADDR_PRIMARY")
	if addr != "" {
		os.Unsetenv("DB_ADDR_PRIMARY")
	}

	dbManager := NewDBManager(s.logger, s.support)
	s.Nil(dbManager.DB("primary"))
	s.Nil(dbManager.DB("mainApp"))
	s.Equal(0, len(dbManager.Errors()))
	s.Equal("* DBs: none", dbManager.Info())
}

func (s *DBManagerSuite) TestNewDBManagerWithCustomConfig() {
	addr := os.Getenv("DB_ADDR_PRIMARY")
	if addr != "" {
		os.Unsetenv("DB_ADDR_PRIMARY")
	}

	defer func() {
		os.Setenv("DB_ADDR_PRIMARY", addr)
	}()

	os.Setenv("DB_SCHEMA_SEARCH_PATH_MAIN_APP", "appist")
	os.Setenv("DB_NETWORK_MAIN_APP", "unix")
	os.Setenv("DB_ADDR_MAIN_APP", "0.0.0.0:25432")
	os.Setenv("DB_USER_MAIN_APP", "appist")
	os.Setenv("DB_PASSWORD_MAIN_APP", "appist")
	os.Setenv("DB_DATABASE_MAIN_APP", "appist")
	os.Setenv("DB_APPLICATION_NAME_MAIN_APP", "appist")
	os.Setenv("DB_REPLICA_MAIN_APP", "true")
	os.Setenv("DB_MAX_RETRIES_MAIN_APP", "3")
	os.Setenv("DB_RETRY_STATEMENT_MAIN_APP", "true")
	os.Setenv("DB_MIN_RETRY_BACKOFF_MAIN_APP", "500ms")
	os.Setenv("DB_MAX_RETRY_BACKOFF_MAIN_APP", "2s")
	os.Setenv("DB_POOL_SIZE_MAIN_APP", "25")
	os.Setenv("DB_POOL_TIMEOUT_MAIN_APP", "25s")
	os.Setenv("DB_MIN_IDLE_CONNS_MAIN_APP", "10")
	os.Setenv("DB_MAX_CONN_AGE_MAIN_APP", "10s")
	os.Setenv("DB_DIAL_TIMEOUT_MAIN_APP", "10s")
	os.Setenv("DB_IDLE_TIMEOUT_MAIN_APP", "25s")
	os.Setenv("DB_IDLE_CHECK_FREQUENCY_MAIN_APP", "2m")
	os.Setenv("DB_READ_TIMEOUT_MAIN_APP", "25s")
	os.Setenv("DB_WRITE_TIMEOUT_MAIN_APP", "25s")
	os.Setenv("DB_SCHEMA_MIGRATIONS_TABLE_MAIN_APP", "custom_migrations")
	defer func() {
		os.Setenv("DB_ADDR_PRIMARY", addr)
		os.Unsetenv("DB_SCHEMA_SEARCH_PATH_MAIN_APP")
		os.Unsetenv("DB_NETWORK_MAIN_APP")
		os.Unsetenv("DB_ADDR_MAIN_APP")
		os.Unsetenv("DB_USER_MAIN_APP")
		os.Unsetenv("DB_PASSWORD_MAIN_APP")
		os.Unsetenv("DB_DATABASE_MAIN_APP")
		os.Unsetenv("DB_APPLICATION_NAME_MAIN_APP")
		os.Unsetenv("DB_REPLICA_MAIN_APP")
		os.Unsetenv("DB_MAX_RETRIES_MAIN_APP")
		os.Unsetenv("DB_RETRY_STATEMENT_MAIN_APP")
		os.Unsetenv("DB_MIN_RETRY_BACKOFF_MAIN_APP")
		os.Unsetenv("DB_MAX_RETRY_BACKOFF_MAIN_APP")
		os.Unsetenv("DB_POOL_SIZE_MAIN_APP")
		os.Unsetenv("DB_POOL_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_MIN_IDLE_CONNS_MAIN_APP")
		os.Unsetenv("DB_MAX_CONN_AGE_MAIN_APP")
		os.Unsetenv("DB_DIAL_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_IDLE_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_IDLE_CHECK_FREQUENCY_MAIN_APP")
		os.Unsetenv("DB_READ_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_WRITE_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_SCHEMA_MIGRATIONS_TABLE_MAIN_APP")
	}()

	dbManager := NewDBManager(s.logger, s.support)
	s.Nil(dbManager.DB("primary"))
	s.NotNil(dbManager.DB("mainApp"))
	s.Equal(0, len(dbManager.Errors()))
	s.Equal("* DBs: mainApp", dbManager.Info())

	config := dbManager.DB("mainApp").Config()
	s.Equal("appist", config.SchemaSearchPath)
	s.Equal("unix", config.Network)
	s.Equal("0.0.0.0:25432", config.Addr)
	s.Equal("appist", config.User)
	s.Equal("appist", config.Password)
	s.Equal("appist", config.Database)
	s.Equal("appist", config.ApplicationName)
	s.Equal(true, config.Replica)
	s.Equal(3, config.MaxRetries)
	s.Equal(true, config.RetryStatementTimeout)
	s.Equal(500*time.Millisecond, config.MinRetryBackoff)
	s.Equal(2*time.Second, config.MaxRetryBackoff)
	s.Equal(25, config.PoolSize)
	s.Equal(25*time.Second, config.PoolTimeout)
	s.Equal(10, config.MinIdleConns)
	s.Equal(10*time.Second, config.MaxConnAge)
	s.Equal(10*time.Second, config.DialTimeout)
	s.Equal(2*time.Minute, config.IdleCheckFrequency)
	s.Equal(25*time.Second, config.IdleTimeout)
	s.Equal(25*time.Second, config.ReadTimeout)
	s.Equal(25*time.Second, config.WriteTimeout)
	s.Equal("custom_migrations", config.SchemaMigrationsTable)
}

func (s *DBManagerSuite) TestNewDBManagerWithInvalidConfig() {
	addr := os.Getenv("DB_ADDR_PRIMARY")
	if addr != "" {
		os.Unsetenv("DB_ADDR_PRIMARY")
	}

	os.Setenv("DB_ADDR_MAIN_APP", "0.0.0.0:25432")
	os.Setenv("DB_REPLICA_MAIN_APP", "100")
	os.Setenv("DB_MAX_RETRIES_MAIN_APP", "true")
	os.Setenv("DB_RETRY_STATEMENT_MAIN_APP", "100")
	os.Setenv("DB_MIN_RETRY_BACKOFF_MAIN_APP", "true")
	os.Setenv("DB_MAX_RETRY_BACKOFF_MAIN_APP", "true")
	os.Setenv("DB_POOL_SIZE_MAIN_APP", "true")
	os.Setenv("DB_POOL_TIMEOUT_MAIN_APP", "true")
	os.Setenv("DB_MIN_IDLE_CONNS_MAIN_APP", "true")
	os.Setenv("DB_MAX_CONN_AGE_MAIN_APP", "true")
	os.Setenv("DB_DIAL_TIMEOUT_MAIN_APP", "true")
	os.Setenv("DB_IDLE_TIMEOUT_MAIN_APP", "true")
	os.Setenv("DB_IDLE_CHECK_FREQUENCY_MAIN_APP", "true")
	os.Setenv("DB_READ_TIMEOUT_MAIN_APP", "true")
	os.Setenv("DB_WRITE_TIMEOUT_MAIN_APP", "true")
	defer func() {
		os.Setenv("DB_ADDR_PRIMARY", addr)
		os.Unsetenv("DB_ADDR_MAIN_APP")
		os.Unsetenv("DB_REPLICA_MAIN_APP")
		os.Unsetenv("DB_MAX_RETRIES_MAIN_APP")
		os.Unsetenv("DB_RETRY_STATEMENT_MAIN_APP")
		os.Unsetenv("DB_MIN_RETRY_BACKOFF_MAIN_APP")
		os.Unsetenv("DB_MAX_RETRY_BACKOFF_MAIN_APP")
		os.Unsetenv("DB_POOL_SIZE_MAIN_APP")
		os.Unsetenv("DB_POOL_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_MIN_IDLE_CONNS_MAIN_APP")
		os.Unsetenv("DB_MAX_CONN_AGE_MAIN_APP")
		os.Unsetenv("DB_DIAL_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_IDLE_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_IDLE_CHECK_FREQUENCY_MAIN_APP")
		os.Unsetenv("DB_READ_TIMEOUT_MAIN_APP")
		os.Unsetenv("DB_WRITE_TIMEOUT_MAIN_APP")
	}()

	dbManager := NewDBManager(s.logger, s.support)
	s.Nil(dbManager.DB("primary"))
	s.NotNil(dbManager.DB("mainApp"))
	s.Equal(14, len(dbManager.Errors()))
	s.Equal("* DBs: mainApp", dbManager.Info())
}

func TestDBManagerSuite(t *testing.T) {
	RunTestSuite(t, new(DBManagerSuite))
}
