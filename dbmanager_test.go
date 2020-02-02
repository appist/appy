package appy

import (
	"testing"
)

type DBManagerSuite struct {
	TestSuite
}

func (s *DBManagerSuite) SetupTest() {
}

func (s *DBManagerSuite) TearDownTest() {
}

func (s *DBManagerSuite) TestParseDBConfig() {
}

func TestDBManagerSuite(t *testing.T) {
	RunTestSuite(t, new(DBManagerSuite))
}
