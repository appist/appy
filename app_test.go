package appy

import (
	"os"
	"testing"
)

type AppSuite struct {
	TestSuite
}

func (s *AppSuite) SetupTest() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	Init(nil, nil)
}

func (s *AppSuite) TearDownTest() {
	os.Clearenv()
}

func (s *AppSuite) TestCmd() {
	s.NotNil(app.Cmd())
}

func (s *AppSuite) TestConfig() {
	s.NotNil(app.Config())
}

func (s *AppSuite) TestDbManager() {
	s.NotNil(app.DbManager())
}

func (s *AppSuite) TestLogger() {
	s.NotNil(app.Logger())
}

func (s *AppSuite) TestServer() {
	s.NotNil(app.Server())
}

func (s *AppSuite) TestRunUnknownCommand() {
	oldArgs := os.Args
	os.Args = append(os.Args, "dummy")
	err := app.Run()
	s.NotNil(err)
	os.Args = oldArgs
}

func TestAppSuite(t *testing.T) {
	RunTestSuite(t, new(AppSuite))
}
