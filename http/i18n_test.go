package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type I18nSuite struct {
	test.Suite
	assets *support.Assets
	config *support.Config
	i18n   *support.I18n
	logger *support.Logger
}

func (s *I18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewFakeLogger()
	layout := map[string]string{
		"docker": "../support/testdata/.docker",
		"config": "../support/testdata/configs",
		"locale": "../support/testdata/pkg/locales",
		"view":   "../support/testdata/pkg/views",
		"web":    "../support/testdata/web",
	}
	s.assets = support.NewAssets(layout, "", http.Dir("../support/testdata"))
	s.config = support.NewConfig(s.assets, s.logger)
	s.i18n = support.NewI18n(s.assets, s.config, s.logger)
}

func (s *I18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *I18nSuite) TestI18n() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: http.Header{
			acceptLanguage: []string{"en-US"},
		},
	}
	I18n(s.i18n)(c)
	s.NotNil(c.i18n)

	val, _ := c.Get(i18nLocaleCtxKey.String())
	s.Equal("en-US", val)
}

func TestI18nSuite(t *testing.T) {
	test.RunSuite(t, new(I18nSuite))
}
