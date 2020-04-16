package pack

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwI18nSuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	i18n   *support.I18n
	logger *support.Logger
}

func (s *mdwI18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "testdata/mdwi18n")
	s.config = support.NewConfig(s.asset, s.logger)
	s.i18n = support.NewI18n(s.asset, s.config, s.logger)
}

func (s *mdwI18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwI18nSuite) TestI18n() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: http.Header{
			acceptLanguage: []string{"en-US"},
		},
	}
	mdwI18n(s.i18n)(c)
	i18n, _ := c.Get(mdwI18nCtxKey.String())
	s.NotNil(i18n)

	locale, _ := c.Get(mdwI18nLocaleCtxKey.String())
	s.Equal("en-US", locale)
}

func TestMdwI18nSuite(t *testing.T) {
	test.Run(t, new(mdwI18nSuite))
}
