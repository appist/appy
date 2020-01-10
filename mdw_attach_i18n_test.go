package appy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type AttachI18nSuite struct {
	TestSuite
	asset  *Asset
	config *Config
	i18n   *I18n
	logger *Logger
}

func (s *AttachI18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewFakeLogger()
	layout := map[string]string{
		"docker": "testdata/i18n/.docker",
		"config": "testdata/i18n/configs",
		"locale": "testdata/i18n/pkg/locales",
		"view":   "testdata/i18n/pkg/views",
		"web":    "testdata/i18n/web",
	}
	s.asset = NewAsset(http.Dir("testdata/i18n"), layout)
	s.config = NewConfig(s.asset, s.logger, &Support{})
	s.i18n = NewI18n(s.asset, s.config, s.logger)
}

func (s *AttachI18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *AttachI18nSuite) TestI18n() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: http.Header{
			acceptLanguage: []string{"en-US"},
		},
	}
	AttachI18n(s.i18n)(c)
	i18n, _ := c.Get(i18nCtxKey.String())
	s.NotNil(i18n)

	locale, _ := c.Get(i18nLocaleCtxKey.String())
	s.Equal("en-US", locale)
}

func TestAttachI18nSuite(t *testing.T) {
	RunTestSuite(t, new(AttachI18nSuite))
}
