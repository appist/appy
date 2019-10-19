package core

import (
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type AppSuite struct {
	test.Suite
}

func (s *AppSuite) SetupTest() {
}

func (s *AppSuite) TearDownTest() {
}

func (s *AppSuite) TestNewApp() {
	oldSSRConfig := SSRPaths["config"]
	SSRPaths["config"] = "./testdata/.ssr/config"
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	app, err := NewApp(nil, nil)
	s.Nil(err)
	s.NotNil(app.Config)
	s.NotNil(app.Logger)
	s.NotNil(app.Server)

	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
	SSRPaths["config"] = oldSSRConfig
}

func (s *AppSuite) TestNewAppWithMissingRequiredEnvVariables() {
	app, err := NewApp(nil, nil)
	s.EqualError(err, "required environment variable \"HTTP_SESSION_SECRETS\" is not set. required environment variable \"HTTP_CSRF_SECRET\" is not set")
	s.Equal(AppConfig{}, app.Config)
}

func TestApp(t *testing.T) {
	test.Run(t, new(AppSuite))
}
