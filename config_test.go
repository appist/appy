package appy_test

import (
	"os"
	"testing"

	"github.com/appist/appy"
)

type ConfigSuite struct {
	appy.TestSuite
}

func (s *ConfigSuite) SetupTest() {
}

func (s *ConfigSuite) TearDownTest() {
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ConfigSuite) TestNewConfigWithoutSettingRequiredConfig() {
	build := "debug"
	support := appy.NewSupport()
	logger := appy.NewLogger(build)
	config := appy.NewConfig(build, logger, support)
	s.NotNil(config.Errors())
	s.EqualError(config.Errors()[0], `required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`)
}

func (s *ConfigSuite) TestNewConfigWithSettingRequiredConfig() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	build := "debug"
	support := appy.NewSupport()
	logger := appy.NewLogger(build)
	config := appy.NewConfig(build, logger, support)
	s.Nil(config.Errors())
}

func TestConfigSuite(t *testing.T) {
	appy.RunTestSuite(t, new(ConfigSuite))
}
