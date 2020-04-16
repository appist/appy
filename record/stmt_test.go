package record

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type stmtSuite struct {
	test.Suite
	buffer    *bytes.Buffer
	db        DBer
	dbManager *Engine
	logger    *support.Logger
	writer    *bufio.Writer
}

func (s *stmtSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *stmtSuite) TearDownTest() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *stmtSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	case "postgres":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
);
`
	}

	s.dbManager = NewEngine(s.logger)
	s.db = s.dbManager.DB("primary")

	err := s.db.DropDB(database)
	s.Nil(err)

	err = s.db.CreateDB(database)
	s.Nil(err)

	err = s.db.Connect()
	s.Nil(err)

	_, err = s.db.Exec(query)
	s.Nil(err)
}

func (s *stmtSuite) TestExec() {
	for _, adapter := range supportedAdapters {
		s.setupDB(adapter, "test_stmt_exec")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		insertStmt, err := s.db.Prepare(query)
		s.Nil(err)

		_, err = insertStmt.Exec("John Doe")
		s.Nil(err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = insertStmt.ExecContext(ctx, "Smith")
		s.Equal("context canceled", err.Error())

		countStmt, err := s.db.Prepare("SELECT COUNT(*) FROM users;")
		s.Nil(err)

		var count int
		err = countStmt.Get(&count)
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *stmtSuite) TestQuery() {
	for _, adapter := range supportedAdapters {
		s.setupDB(adapter, "test_stmt_query")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		insertStmt, err := s.db.Prepare(query)
		s.Nil(err)

		_, err = insertStmt.Query("John Doe")
		s.Nil(err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = insertStmt.QueryContext(ctx, "Smith")
		s.Equal("context canceled", err.Error())

		countStmt, err := s.db.Prepare("SELECT COUNT(*) FROM users;")
		s.Nil(err)

		var count int
		err = countStmt.Get(&count)
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *stmtSuite) TestQueryRow() {
	for _, adapter := range supportedAdapters {
		s.setupDB(adapter, "test_stmt_query_row")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		insertStmt, err := s.db.Prepare(query)
		s.Nil(err)

		_ = insertStmt.QueryRow("John Doe")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_ = insertStmt.QueryRowContext(ctx, "Smith")

		countStmt, err := s.db.Prepare("SELECT COUNT(*) FROM users;")
		s.Nil(err)

		var count int
		err = countStmt.Get(&count)
		s.Nil(err)
		s.Equal(1, count)
	}
}

func TestStmtSuite(t *testing.T) {
	test.Run(t, new(stmtSuite))
}
