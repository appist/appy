package appy

import (
	"os"
	"testing"
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

func (s *DBManagerSuite) TestDBManagerWithBadDBConfig() {
	os.Setenv("DB_URI_PRIMARY", "0.0.0.0:13306/appy")
	defer os.Unsetenv("DB_URI_PRIMARY")
	dbManager := NewDBManager(s.logger, s.support)
	s.Equal(1, len(dbManager.Errors()))
}

func (s *DBManagerSuite) TestDBManagerWithValidDBConfig() {
	os.Setenv("DB_URI_PRIMARY", "mysql://root:whatever@0.0.0.0:13306/appy")
	os.Setenv("DB_URI_PRIMARY_REPLICA", "mysql://root:whatever@0.0.0.0:13307/appy")
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_URI_PRIMARY_REPLICA")
	}()
	dbManager := NewDBManager(s.logger, s.support)
	s.Equal(0, len(dbManager.Errors()))
	s.Nil(dbManager.DB("foobar"))
	s.NotNil(dbManager.DB("primary"))
	s.NotNil(dbManager.DB("primaryReplica"))
	s.Contains(dbManager.Info(), "* DBs:")
	s.Contains(dbManager.Info(), "primary")
	s.Contains(dbManager.Info(), "primaryReplica")
}

func TestDBManagerSuite(t *testing.T) {
	RunTestSuite(t, new(DBManagerSuite))
}
