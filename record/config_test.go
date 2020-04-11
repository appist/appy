package record

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type configSuite struct {
	test.Suite
	buffer *bytes.Buffer
	writer *bufio.Writer
	logger *support.Logger
}

func (s *configSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *configSuite) TestParseDBConfigWithBadConfig() {
	dbConfigs, errs := parseDBConfig()
	s.Nil(errs)
	s.Equal(0, len(dbConfigs))

	os.Setenv("DB_URI_PRIMARY", "0.0.0.0:13306/appy")
	dbConfigs, errs = parseDBConfig()
	s.Equal(1, len(errs))
	s.Equal(0, len(dbConfigs))

	os.Setenv("DB_CONN_MAX_LIFETIME_PRIMARY", "true")
	os.Setenv("DB_MAX_IDLE_CONNS_PRIMARY", "true")
	os.Setenv("DB_MAX_OPEN_CONNS_PRIMARY", "true")
	os.Setenv("DB_REPLICA_PRIMARY", "10s")
	os.Setenv("DB_URI_PRIMARY", "sqlite://root:whatever@0.0.0.0:13306/appy")
	defer func() {
		os.Unsetenv("DB_CONN_MAX_LIFETIME_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
		os.Unsetenv("DB_REPLICA_PRIMARY")
		os.Unsetenv("DB_URI_PRIMARY")
	}()

	dbConfigs, errs = parseDBConfig()
	s.Equal(5, len(errs))
	s.Equal(0, len(dbConfigs))
}

func (s *configSuite) TestParseDBConfigWithMultipleURIs() {
	os.Setenv("DB_URI_PRIMARY", "mysql://root:whatever@0.0.0.0:13306/appy")
	os.Setenv("DB_URI_SECONDARY", "mysql://root:whatever@0.0.0.0:13307/appist")
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_URI_SECONDARY")
	}()

	dbConfigs, errs := parseDBConfig()
	s.Nil(errs)
	s.Equal(2, len(dbConfigs))
}

func (s *configSuite) TestParseDBConfigForMySQL() {
	os.Setenv("DB_CONN_MAX_LIFETIME_PRIMARY", "10m")
	os.Setenv("DB_MAX_IDLE_CONNS_PRIMARY", "50")
	os.Setenv("DB_MAX_OPEN_CONNS_PRIMARY", "100")
	os.Setenv("DB_REPLICA_PRIMARY", "true")
	os.Setenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY", "mysql_migrations")
	os.Setenv("DB_URI_PRIMARY", "mysql://root:whatever@0.0.0.0:13306/appy")
	defer func() {
		os.Unsetenv("DB_CONN_MAX_LIFETIME_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
		os.Unsetenv("DB_REPLICA_PRIMARY")
		os.Unsetenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY")
		os.Unsetenv("DB_URI_PRIMARY")
	}()

	dbConfigs, errs := parseDBConfig()
	s.Nil(errs)
	s.Equal(1, len(dbConfigs))

	dbConfig := dbConfigs["primary"]
	s.Equal("mysql", dbConfig.Adapter)
	s.Equal(10*time.Minute, dbConfig.ConnMaxLifetime)
	s.Equal("appy", dbConfig.Database)
	s.Equal("0.0.0.0", dbConfig.Host)
	s.Equal(50, dbConfig.MaxIdleConns)
	s.Equal(100, dbConfig.MaxOpenConns)
	s.Equal("whatever", dbConfig.Password)
	s.Equal("13306", dbConfig.Port)
	s.Equal(true, dbConfig.Replica)
	s.Equal("mysql_migrations", dbConfig.SchemaMigrationsTable)
	s.Equal("root:whatever@tcp(0.0.0.0:13306)/appy", dbConfig.URI)
	s.Equal("root", dbConfig.Username)
}

func (s *configSuite) TestParseDBConfigForPostgreSQL() {
	os.Setenv("DB_CONN_MAX_LIFETIME_PRIMARY", "10m")
	os.Setenv("DB_MAX_IDLE_CONNS_PRIMARY", "50")
	os.Setenv("DB_MAX_OPEN_CONNS_PRIMARY", "100")
	os.Setenv("DB_REPLICA_PRIMARY", "true")
	os.Setenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY", "postgres_migrations")
	os.Setenv("DB_SCHEMA_SEARCH_PATH_PRIMARY", "public,appy")
	os.Setenv("DB_URI_PRIMARY", "postgresql://postgres:whatever@0.0.0.0:15432/appy?sslmode=disable&application_name=appy&connect_timeout=5")
	defer func() {
		os.Unsetenv("DB_CONN_MAX_LIFETIME_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
		os.Unsetenv("DB_REPLICA_PRIMARY")
		os.Unsetenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY")
		os.Unsetenv("DB_SCHEMA_SEARCH_PATH_PRIMARY")
		os.Unsetenv("DB_URI_PRIMARY")
	}()

	dbConfigs, errs := parseDBConfig()
	s.Nil(errs)
	s.Equal(1, len(dbConfigs))

	dbConfig := dbConfigs["primary"]
	s.Equal("postgres", dbConfig.Adapter)
	s.Equal(10*time.Minute, dbConfig.ConnMaxLifetime)
	s.Equal("appy", dbConfig.Database)
	s.Equal("0.0.0.0", dbConfig.Host)
	s.Equal(50, dbConfig.MaxIdleConns)
	s.Equal(100, dbConfig.MaxOpenConns)
	s.Equal("whatever", dbConfig.Password)
	s.Equal("15432", dbConfig.Port)
	s.Equal(true, dbConfig.Replica)
	s.Equal("postgres_migrations", dbConfig.SchemaMigrationsTable)
	s.Equal("public,appy", dbConfig.SchemaSearchPath)
	s.Equal("postgresql://postgres:whatever@0.0.0.0:15432/appy?sslmode=disable&application_name=appy&connect_timeout=5", dbConfig.URI)
	s.Equal("postgres", dbConfig.Username)
}

func TestConfigSuite(t *testing.T) {
	test.Run(t, new(configSuite))
}
