package orm

import (
	"net"
	"os"
	"testing"

	"github.com/appist/appy/internal/test"
)

type DbManagerSuite struct {
	test.Suite
}

func (s *DbManagerSuite) SetupTest() {
}

func (s *DbManagerSuite) TearDownTest() {
}

func (s *DbManagerSuite) TestParseInvalidDbConfig() {
	os.Setenv("DB_ADDR_PRIMARY", "dummy")
	os.Setenv("DB_APP_NAME_PRIMARY", "dummy")
	os.Setenv("DB_DIAL_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_IDLE_CHECK_FREQUENCY_PRIMARY", "dummy")
	os.Setenv("DB_IDLE_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_MAX_CONN_AGE_PRIMARY", "dummy")
	os.Setenv("DB_MAX_RETRIES_PRIMARY", "dummy")
	os.Setenv("DB_MIN_IDLE_CONNS_PRIMARY", "dummy")
	os.Setenv("DB_NAME_PRIMARY", "dummy")
	os.Setenv("DB_PASSWORD_PRIMARY", "dummy")
	os.Setenv("DB_POOL_SIZE_PRIMARY", "dummy")
	os.Setenv("DB_POOL_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_READ_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_REPLICA_PRIMARY", "dummy")
	os.Setenv("DB_RETRY_STATEMENT_PRIMARY", "dummy")
	os.Setenv("DB_SCHEMA_SEARCH_PATH_PRIMARY", "dummy")
	os.Setenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY", "true")
	os.Setenv("DB_USER_PRIMARY", "dummy")
	os.Setenv("DB_WRITE_TIMEOUT_PRIMARY", "dummy")
	defer func() {
		os.Unsetenv("DB_ADDR_PRIMARY")
		os.Unsetenv("DB_APP_NAME_PRIMARY")
		os.Unsetenv("DB_DIAL_TIMEOUT_PRIMARY")
		os.Unsetenv("DB_IDLE_CHECK_FREQUENCY_PRIMARY")
		os.Unsetenv("DB_IDLE_TIMEOUT_PRIMARY")
		os.Unsetenv("DB_MAX_CONN_AGE_PRIMARY")
		os.Unsetenv("DB_MAX_RETRIES_PRIMARY")
		os.Unsetenv("DB_MIN_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_NAME_PRIMARY")
		os.Unsetenv("DB_PASSWORD_PRIMARY")
		os.Unsetenv("DB_POOL_SIZE_PRIMARY")
		os.Unsetenv("DB_POOL_TIMEOUT_PRIMARY")
		os.Unsetenv("DB_READ_TIMEOUT_PRIMARY")
		os.Unsetenv("DB_REPLICA_PRIMARY")
		os.Unsetenv("DB_RETRY_STATEMENT_PRIMARY")
		os.Unsetenv("DB_SCHEMA_SEARCH_PATH_PRIMARY")
		os.Unsetenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY")
		os.Unsetenv("DB_USER_PRIMARY")
		os.Unsetenv("DB_WRITE_TIMEOUT_PRIMARY")
	}()

	_, errs := parseDbConfig()
	s.NotNil(errs)
}

func (s *DbManagerSuite) TestParseValidDbConfig() {
	// A workaround for Github Action
	_, err := net.Dial("tcp", "0.0.0.0:5432")
	if err != nil {
		os.Setenv("DB_ADDR_PRIMARY", "localhost:32768")
	}

	_, errs := parseDbConfig()
	s.Nil(errs)
}

func TestDbManagerSuite(t *testing.T) {
	test.RunSuite(t, new(DbManagerSuite))
}
