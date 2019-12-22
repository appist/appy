package support

import (
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type I18nSuite struct {
	test.Suite
	assets *Assets
	config *Config
	logger *Logger
}

func (s *I18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewFakeLogger()
	layout := map[string]string{
		"docker": "../support/testdata/.docker",
		"config": "../support/testdata/configs",
		"locale": "../support/testdata/pkg/locales",
		"view":   "../support/testdata/pkg/views",
		"web":    "../support/testdata/web",
	}
	s.assets = NewAssets(layout, "", http.Dir("../support/testdata"))
	s.config = NewConfig(s.assets, s.logger)
}

func (s *I18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *I18nSuite) TestInitializePanic() {
	s.assets = NewAssets(nil, "", http.Dir("../support/testdata"))
	s.Panics(func() { NewI18n(s.assets, s.config, s.logger) })
}

func (s *I18nSuite) TestBundle() {
	i18n := NewI18n(s.assets, s.config, s.logger)
	s.NotNil(i18n.Bundle())
}

func (s *I18nSuite) TestLocales() {
	i18n := NewI18n(s.assets, s.config, s.logger)
	s.Equal([]string{"en", "zh-CN", "zh-TW"}, i18n.Locales())
}

func (s *I18nSuite) TestT() {
	i18n := NewI18n(s.assets, s.config, s.logger)
	s.Equal("Test", i18n.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", i18n.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", i18n.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", i18n.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", i18n.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", i18n.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", i18n.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", i18n.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))

	s.Equal("", i18n.T("title.foo", "en"))
}

func TestI18nSuite(t *testing.T) {
	test.RunSuite(t, new(I18nSuite))
}
