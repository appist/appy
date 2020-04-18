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

type txSuite struct {
	test.Suite
	buffer    *bytes.Buffer
	db        DBer
	dbManager *Engine
	logger    *support.Logger
	writer    *bufio.Writer
}

func (s *txSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *txSuite) TearDownTest() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *txSuite) setupDB(adapter, database string) {
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

func (s *txSuite) TestExec() {
	var count int

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_tx_exec")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		tx, err := s.db.Begin()
		s.Nil(err)

		_, err = tx.Exec(query, "John Doe")
		s.Nil(err)

		_, err = tx.ExecContext(ctx, query, "John Doe")
		s.Equal("context canceled", err.Error())

		err = tx.Commit()
		s.Nil(err)

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *txSuite) TestPrepare() {
	var count int

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_tx_prepare")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		tx, err := s.db.Begin()
		s.Nil(err)

		stmt, err := tx.Prepare(query)
		s.Nil(err)

		_, err = stmt.Exec("John Doe")
		s.Nil(err)

		_, err = tx.PrepareContext(ctx, query)
		s.Equal("context canceled", err.Error())

		err = tx.Commit()
		s.Nil(err)

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *txSuite) TestQuery() {
	var count int

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_tx_query")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		tx, err := s.db.Begin()
		s.Nil(err)

		rows, err := tx.Query(query, "John Doe")
		rows.Close()
		s.Nil(err)

		_, err = tx.QueryContext(ctx, query, "John Doe")
		s.Equal("context canceled", err.Error())

		err = tx.Commit()
		s.Nil(err)

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *txSuite) TestQueryRow() {
	var count int

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_tx_query_row")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		tx, err := s.db.Begin()
		s.Nil(err)

		row := tx.QueryRow(query, "John Doe")
		row.Scan()

		_ = tx.QueryRowContext(ctx, query, "John Doe")

		err = tx.Commit()
		s.Nil(err)

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *txSuite) TestStmt() {
	var count int

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_tx_stmt")

		query := `INSERT INTO users (username) VALUES (?);`
		if adapter == "postgres" {
			query = `INSERT INTO users (username) VALUES ($1);`
		}

		stmt, err := s.db.Prepare(query)
		s.Nil(err)

		tx, err := s.db.Begin()
		s.Nil(err)

		_, err = tx.Stmt(stmt).Exec("John Doe")
		s.Nil(err)

		_, err = tx.StmtContext(ctx, stmt).Exec("John Doe")
		s.Equal("context canceled", err.Error())

		err = tx.Commit()
		s.Nil(err)

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(1, count)
	}
}

func TestTxSuite(t *testing.T) {
	test.Run(t, new(txSuite))
}
