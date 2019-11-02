package appy_test

import (
	"os"
	"testing"

	"github.com/appist/appy"
)

type AppSuite struct {
	appy.TestSuite
	app *appy.App
}

func (s *AppSuite) SetupTest() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.app = appy.NewApp(nil, nil)
}

func (s *AppSuite) TearDownTest() {
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *AppSuite) TestAppCmd() {
	s.NotNil(s.app.Cmd())
}

func (s *AppSuite) TestAppConfig() {
	s.NotNil(s.app.Config())
}

func (s *AppSuite) TestAppDbManager() {
}

func (s *AppSuite) TestAppLogger() {
	s.NotNil(s.app.Logger())
}

func (s *AppSuite) TestAppServer() {
}

func TestAppSuite(t *testing.T) {
	appy.RunTestSuite(t, new(AppSuite))
}
