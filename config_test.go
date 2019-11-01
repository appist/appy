package appy

import (
	"os"
	"testing"
)

type ConfigSuite struct {
	TestSuite
	oldSSRPaths map[string]string
}

func (s *ConfigSuite) SetupTest() {
	s.oldSSRPaths = _ssrPaths
	_ssrPaths = map[string]string{
		"root":   "testdata/.ssr",
		"config": "testdata/pkg/config",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
	}
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
}

func (s *ConfigSuite) TearDownTest() {
	_ssrPaths = s.oldSSRPaths
	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithoutSettingRequiredConfig() {
	build := DebugBuild
	support := NewSupport()
	logger := NewLogger(build)
	config := NewConfig(build, logger, support, nil)
	s.NotNil(config.Errors())
	s.EqualError(config.Errors()[0], `required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`)
}

func (s *ConfigSuite) TestNewConfigWithSettingRequiredConfig() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	support := NewSupport()
	logger := NewLogger(build)
	config := NewConfig(build, logger, support, nil)
	s.Nil(config.Errors())

	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func TestConfigSuite(t *testing.T) {
	RunTestSuite(t, new(ConfigSuite))
}
