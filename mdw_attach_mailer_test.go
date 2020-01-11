package appy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type AttachMailerSuite struct {
	TestSuite
	asset   *Asset
	config  *Config
	i18n    *I18n
	logger  *Logger
	mailer  *Mailer
	server  *Server
	support Supporter
}

func (s *AttachMailerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, _, _ = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata/app"), map[string]string{
		"docker": "testdata/app/.docker",
		"config": "testdata/app/configs",
		"locale": "testdata/app/pkg/locales",
		"view":   "testdata/app/pkg/views",
		"web":    "testdata/app/web",
	})
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.i18n = NewI18n(s.asset, s.config, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger, s.support)
	s.mailer = NewMailer(s.asset, s.config, s.i18n, s.logger, s.server, nil)
}

func (s *AttachMailerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *AttachMailerSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	AttachMailer(s.mailer)(c)
	s.NotNil(c.Get(mailerCtxKey.String()))
}

func TestAttachMailerSuite(t *testing.T) {
	RunTestSuite(t, new(AttachMailerSuite))
}
