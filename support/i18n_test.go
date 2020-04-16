package support

import (
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type i18nSuite struct {
	test.Suite
	asset  *Asset
	config *Config
	logger *Logger
}

func (s *i18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewTestLogger()
}

func (s *i18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *i18nSuite) TestMissingLocales() {
	s.asset = NewAsset(nil, "testdata/missing")
	s.config = NewConfig(s.asset, s.logger)

	s.Panics(func() { NewI18n(s.asset, s.config, s.logger) })
}

func (s *i18nSuite) TestTWithDebugBuild() {
	s.asset = NewAsset(nil, "testdata/i18n/t_with_debug_build")
	s.config = NewConfig(s.asset, s.logger)
	i18n := NewI18n(s.asset, s.config, s.logger)

	s.NotNil(i18n.Bundle())
	s.ElementsMatch([]string{"en", "zh-TW", "zh-CN"}, i18n.Locales())
	s.Equal("", i18n.T("title.foo", "en"))

	s.Equal("Test", i18n.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", i18n.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", i18n.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", i18n.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", i18n.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", i18n.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", i18n.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", i18n.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))
}

func (s *i18nSuite) TestTWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	s.asset = NewAsset(http.Dir("testdata/i18n/t_with_release_build"), "")
	s.config = NewConfig(s.asset, s.logger)
	i18n := NewI18n(s.asset, s.config, s.logger)

	s.NotNil(i18n.Bundle())
	s.ElementsMatch([]string{"en", "zh-TW", "zh-CN"}, i18n.Locales())
	s.Equal("", i18n.T("title.foo", "en"))

	s.Equal("Test", i18n.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", i18n.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", i18n.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", i18n.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", i18n.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", i18n.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", i18n.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", i18n.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))
}

func TestI18nSuite(t *testing.T) {
	test.Run(t, new(i18nSuite))
}
