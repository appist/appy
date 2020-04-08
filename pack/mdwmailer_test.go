package pack

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwMailerSuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	i18n   *support.I18n
	logger *support.Logger
	mailer *mailer.Engine
	server *Server
}

func (s *mdwMailerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "testdata/context")
	s.config = support.NewConfig(s.asset, s.logger)
	s.i18n = support.NewI18n(s.asset, s.config, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger)
	s.mailer = mailer.NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
}

func (s *mdwMailerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwMailerSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	mdwMailer(s.mailer)(c)

	s.NotNil(c.Get(mdwMailerCtxKey.String()))
}

func TestMdwMailerSuite(t *testing.T) {
	test.Run(t, new(mdwMailerSuite))
}
