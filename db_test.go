package appy

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"
)

type DBSuite struct {
	TestSuite
	buffer          *bytes.Buffer
	dbManager       *DBManager
	writer          *bufio.Writer
	logger          *Logger
	support         Supporter
	mysqlDB, psqlDB DBer
}

func (s *DBSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.support = &Support{}
}

func (s *DBSuite) setupDB(database string) {
	os.Setenv("DB_URI_MYSQL", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	os.Setenv("DB_URI_POSTGRES", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
	s.dbManager = NewDBManager(s.logger, s.support)

	s.mysqlDB = s.dbManager.DB("mysql")
	s.mysqlDB.ConnectDB("mysql")
	s.mysqlDB.DropDB(database)
	s.mysqlDB.CreateDB(database)
	s.mysqlDB.Connect()
	s.mysqlDB.Exec(`
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`)

	s.psqlDB = s.dbManager.DB("postgres")
	s.psqlDB.ConnectDB("postgres")
	s.psqlDB.DropDB(database)
	s.psqlDB.CreateDB(database)
	s.psqlDB.Connect()
	s.psqlDB.Exec(`
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
);
`)
}

func (s *DBSuite) teardownDB(database string) {
	os.Unsetenv("DB_URI_MYSQL")
	os.Unsetenv("DB_URI_POSTGRES")
}

func (s *DBSuite) TestTransactionForMySQL() {
	database := "test_transaction_for_mysql"
	s.setupDB(database)
	defer s.teardownDB(database)

	tx, err := s.mysqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES(?);", "foobar")
	s.Nil(err)
	tx.Rollback()

	var count int
	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)

	tx, err = s.mysqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES(?);", "foobar")
	s.Nil(err)
	tx.Commit()

	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)
}

func (s *DBSuite) TestTransactionForPostgreSQL() {
	database := "test_transaction_for_postgresql"
	s.setupDB(database)
	defer s.teardownDB(database)

	tx, err := s.psqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES($1);", "foobar")
	s.Nil(err)
	tx.Rollback()

	var count int
	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)

	tx, err = s.psqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES($1);", "foobar")
	s.Nil(err)
	tx.Commit()

	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)
}

func TestDBSuite(t *testing.T) {
	RunTestSuite(t, new(DBSuite))
}
