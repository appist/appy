package core

import (
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type AppSuite struct {
	test.Suite
	oldConfigPath string
}

func (s *AppSuite) SetupTest() {
}

func (s *AppSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *AppSuite) TestNewApp() {
	os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")

	app, err := NewApp(nil, nil)
	s.Nil(err)
	s.NotNil(app.Config)
	s.NotNil(app.Logger)
	s.NotNil(app.Server)
}

func (s *AppSuite) TestNewAppWithMissingRequiredEnvVariables() {
	app, err := NewApp(nil, nil)
	s.EqualError(err, "required environment variable \"HTTP_SESSION_SECRETS\" is not set. required environment variable \"HTTP_CSRF_SECRET\" is not set")
	s.Equal(AppConfig{}, app.Config)
}

func TestApp(t *testing.T) {
	test.Run(t, new(AppSuite))
}
