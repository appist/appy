package appy

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type DBSuite struct {
	TestSuite
	buffer  *bytes.Buffer
	writer  *bufio.Writer
	logger  *Logger
	support Supporter
}

func (s *DBSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.logger.SetDBLogging(true)
	s.support = &Support{}
}

func (s *DBSuite) TestDBGenerateMigration() {
	os.Setenv("DB_ADDR_PRIMARY", "0.0.0.0:5432")
	os.Setenv("DB_USER_PRIMARY", "postgres")
	os.Setenv("DB_PASSWORD_PRIMARY", "whatever")
	os.Setenv("DB_DATABASE_PRIMARY", "appy")
	oldDBMigratePath := dbMigratePath
	dbMigratePath = "tmp/" + dbMigratePath
	defer func() {
		dbMigratePath = oldDBMigratePath
		os.Unsetenv("DB_ADDR_PRIMARY")
		os.Unsetenv("DB_USER_PRIMARY")
		os.Unsetenv("DB_PASSWORD_PRIMARY")
		os.Unsetenv("DB_DATABASE_PRIMARY")
	}()

	config, errs := parseDBConfig(s.support)
	s.Nil(errs)

	migratePath := dbMigratePath + "primary"
	db := NewDB(config["primary"], s.logger, s.support)
	db.GenerateMigration("CreateUsers", "primary", false)
	files := []string{}
	_ = filepath.Walk(migratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err := ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigration(")
	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)

	db.GenerateMigration("CreateUsers", "primary", true)
	files = []string{}
	_ = filepath.Walk(migratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err = ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigrationTx(")
	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)
}

func (s *DBSuite) TestDBOps() {
	if os.Getenv("DB_ADDR_PRIMARY") == "" {
		os.Setenv("DB_ADDR_PRIMARY", "0.0.0.0:5432")
	}

	os.Setenv("DB_USER_PRIMARY", "postgres")
	os.Setenv("DB_PASSWORD_PRIMARY", "whatever")
	os.Setenv("DB_DATABASE_PRIMARY", "appy")
	oldDBMigratePath := dbMigratePath
	dbMigratePath = "tmp/" + dbMigratePath
	defer func() {
		dbMigratePath = oldDBMigratePath
		os.Unsetenv("DB_ADDR_PRIMARY")
		os.Unsetenv("DB_USER_PRIMARY")
		os.Unsetenv("DB_PASSWORD_PRIMARY")
		os.Unsetenv("DB_DATABASE_PRIMARY")
	}()

	config, errs := parseDBConfig(s.support)
	s.Nil(errs)

	db := NewDB(config["primary"], s.logger, s.support)

	// Test DB create
	targetDB := db.Config().Database
	db.config.Database = "postgres"
	err := db.Connect()
	defer db.Close()
	s.writer.Flush()
	s.Nil(err)
	s.Equal("", s.buffer.String())

	db.config.Database = targetDB
	errs = db.Create()
	s.writer.Flush()
	s.Nil(errs)
	s.Contains(s.buffer.String(), "CREATE DATABASE appy")
	s.Contains(s.buffer.String(), "CREATE DATABASE appy_test")

	res, err := db.Exec("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower(?);", "appy")
	s.Nil(err)
	s.Equal(1, res.RowsReturned())

	res, err = db.Exec("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower(?);", "appy_test")
	s.Nil(err)
	s.Equal(1, res.RowsReturned())

	errs = db.Create()
	s.NotNil(errs)

	// Test DB Migrate
	err = db.Connect()
	s.Nil(err)

	db.RegisterMigrationTx(
		func(h *DBTx) error {
			_, err := h.Exec(`
					CREATE TABLE IF NOT EXISTS users (
						id SERIAL PRIMARY KEY,
						confirmation_token VARCHAR,
						confirmation_sent_at TIMESTAMP,
						confirmed_at TIMESTAMP,
						email VARCHAR UNIQUE NOT NULL,
						username VARCHAR UNIQUE NOT NULL,
						encrypted_password VARCHAR(128),
						failed_attempts INT4 NOT NULL DEFAULT 0,
						locked_at TIMESTAMP,
						unlock_token VARCHAR,
						reset_password_sent_at TIMESTAMP,
						reset_password_token VARCHAR,
						created_at TIMESTAMP NOT NULL,
						deleted_at TIMESTAMP,
						updated_at TIMESTAMP
					);
				`)

			return err
		},
		func(h *DBTx) error {
			_, err := h.Exec(`DROP TABLE IF EXISTS users;`)

			return err
		},
		"20200201165238_create_users",
	)

	db.RegisterMigration(
		func(h *DB) error {
			_, err := h.Exec(`CREATE INDEX CONCURRENTLY users_on_deleted_at ON users (deleted_at);`)
			return err
		},
		func(h *DB) error {
			_, err := h.Exec(`DROP INDEX users_on_deleted_at;`)
			return err
		},
		"20200202165238_add_users_on_deleted_at_index",
	)

	err = db.Migrate()
	s.Nil(err)

	migrations, err := db.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("up", migrations[1][0])

	// Test DB dump schema
	err = db.DumpSchema("appy")
	s.Nil(err)

	schemaPath := dbMigratePath + "/appy/schema.go"
	_, err = os.Stat(schemaPath)
	s.Equal(false, os.IsNotExist(err))
	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)

	// Test DB rollback
	err = db.Rollback()
	s.Nil(err)

	migrations, err = db.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("down", migrations[1][0])

	err = db.Rollback()
	s.Nil(err)

	migrations, err = db.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("down", migrations[0][0])
	s.Equal("down", migrations[1][0])

	// Test DB drop
	targetDB = db.Config().Database
	db.config.Database = "postgres"
	_ = db.Connect()
	defer db.Close()
	s.Nil(err)

	db.config.Database = targetDB
	errs = db.Drop()
	s.writer.Flush()
	s.Nil(errs)
	s.Contains(s.buffer.String(), "DROP DATABASE appy")
	s.Contains(s.buffer.String(), "DROP DATABASE appy_test")

	res, err = db.Exec("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower(?);", "appy")
	s.Nil(err)
	s.Equal(0, res.RowsReturned())

	res, err = db.Exec("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower(?);", "appy_test")
	s.Nil(err)
	s.Equal(0, res.RowsReturned())

	errs = db.Drop()
	s.NotNil(errs)
}

func (s *DBSuite) TestDBSchema() {
	os.Setenv("DB_ADDR_PRIMARY", "0.0.0.0:5432")
	os.Setenv("DB_USER_PRIMARY", "postgres")
	os.Setenv("DB_PASSWORD_PRIMARY", "whatever")
	os.Setenv("DB_DATABASE_PRIMARY", "appy")
	defer func() {
		os.Unsetenv("DB_ADDR_PRIMARY")
		os.Unsetenv("DB_USER_PRIMARY")
		os.Unsetenv("DB_PASSWORD_PRIMARY")
		os.Unsetenv("DB_DATABASE_PRIMARY")
	}()

	config, errs := parseDBConfig(s.support)
	s.Nil(errs)

	db := NewDB(config["primary"], s.logger, s.support)
	s.Equal("", db.Schema())

	db.SetSchema("SQL schema")
	s.Equal("SQL schema", db.Schema())
}

func TestDBSuite(t *testing.T) {
	RunTestSuite(t, new(DBSuite))
}
