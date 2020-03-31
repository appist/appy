package appy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
)

type DBTxSuite struct {
	TestSuite
	buffer    *bytes.Buffer
	dbManager *DBManager
	writer    *bufio.Writer
	logger    *Logger
	support   Supporter
	db        DBer
}

func (s *DBTxSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.support = &Support{}
}

func (s *DBTxSuite) setupDB(database string) {
	os.Setenv("DB_URI_MYSQL", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	defer os.Unsetenv("DB_URI_MYSQL")

	s.dbManager = NewDBManager(s.logger, s.support)
	s.db = s.dbManager.DB("mysql")
	s.Equal("0.0.0.0", s.db.Config().Host)
	s.Equal("13306", s.db.Config().Port)
	s.Equal(database, s.db.Config().Database)

	s.db.ConnectDB("mysql")
	s.db.DropDB(database)
	s.db.CreateDB(database)
	s.db.Connect()
	s.db.Exec(`
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`)
}

func (s *DBTxSuite) TestOperations() {
	database := "test_dbtx_operations"
	s.setupDB(database)

	err := s.db.Connect()
	s.Nil(err)

	tx, err := s.db.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users (username) VALUES (?);", "Smith")
	s.Nil(err)
	s.Nil(tx.Commit())

	var count int
	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)

	tx, err = s.db.Begin()
	s.Nil(err)

	stmt, err := tx.Prepare(`INSERT INTO users (username) VALUES (?);`)
	s.Nil(err)

	_ = stmt.QueryRow("Smith")
	s.Nil(tx.Commit())

	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(2, count)

	tx, err = s.db.Begin()
	s.Nil(err)

	_, err = tx.Query(`INSERT INTO users (username) VALUES (?);`, "Smith")
	s.Nil(err)
	s.Nil(tx.Commit())

	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(3, count)

	tx, err = s.db.Begin()
	s.Nil(err)

	_ = tx.QueryRow(`INSERT INTO users (username) VALUES (?);`, "Smith")
	s.Nil(tx.Commit())

	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(4, count)

	tx, err = s.db.Begin()
	s.Nil(err)

	query, err := s.db.Prepare(`INSERT INTO users (username) VALUES (?);`)
	s.Nil(err)

	_, err = tx.Stmt(query).Exec("Smith")
	s.Nil(tx.Commit())

	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(5, count)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tx, err = s.db.Begin()
	s.Nil(err)

	_, err = tx.ExecContext(ctx, `INSERT INTO users (username) VALUES (?);`, "Smith")
	s.Equal("context canceled", err.Error())

	_, err = tx.PrepareContext(ctx, `INSERT INTO users (username) VALUES (?);`)
	s.Equal("context canceled", err.Error())

	_, err = tx.QueryContext(ctx, `INSERT INTO users (username) VALUES (?);`, "Smith")
	s.Equal("context canceled", err.Error())

	_ = tx.QueryRowContext(ctx, `INSERT INTO users (username) VALUES (?);`, "Smith")
	s.Equal(nil, tx.Commit())

	err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(5, count)

	tx, err = s.db.Begin()
	s.Nil(err)

	_, err = tx.StmtContext(ctx, query).Exec("Smith")
	s.Equal("context canceled", err.Error())
}

func TestDBTxSuite(t *testing.T) {
	RunTestSuite(t, new(DBTxSuite))
}
