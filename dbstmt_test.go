package appy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
)

type DBStmtSuite struct {
	TestSuite
	buffer    *bytes.Buffer
	dbManager *DBManager
	writer    *bufio.Writer
	logger    *Logger
	support   Supporter
	db        DBer
}

func (s *DBStmtSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.support = &Support{}
}

func (s *DBStmtSuite) setupDB(database string) {
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

func (s *DBStmtSuite) TestOperations() {
	database := "test_dbstmt_operations"
	s.setupDB(database)

	err := s.db.Connect()
	s.Nil(err)

	insertStmt, err := s.db.Prepare(`INSERT INTO users (username) VALUES (?);`)
	s.Nil(err)

	_, err = insertStmt.Exec("Smith")
	s.Nil(err)

	countStmt, err := s.db.Prepare("SELECT COUNT(*) FROM users;")
	s.Nil(err)

	var count int
	err = countStmt.Get(&count)
	s.Nil(err)
	s.Equal(1, count)

	_, err = insertStmt.Query("Smith")
	s.Nil(err)

	err = countStmt.Get(&count)
	s.Nil(err)
	s.Equal(2, count)

	_ = insertStmt.QueryRow("Smith")
	s.Nil(err)

	err = countStmt.Get(&count)
	s.Nil(err)
	s.Equal(3, count)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = insertStmt.ExecContext(ctx, "Smith")
	s.Equal("context canceled", err.Error())

	err = countStmt.GetContext(ctx, &count)
	s.Equal("context canceled", err.Error())

	_, err = countStmt.QueryContext(ctx)
	s.Equal("context canceled", err.Error())

	_ = insertStmt.QueryRowContext(ctx)
	err = countStmt.Get(&count)
	s.Nil(err)
	s.Equal(3, count)
}

func TestDBStmtSuite(t *testing.T) {
	RunTestSuite(t, new(DBStmtSuite))
}
