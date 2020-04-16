package pack

import (
	"net/http"
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
	s.mailer = mailer.NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
	s.server = NewServer(s.asset, s.config, s.logger)
	s.server.Use(mdwLogger(s.logger))
}

func (s *mdwMailerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwMailerSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	mdwMailer(s.mailer, s.i18n, s.server)(c)

	s.NotNil(c.Get(mdwMailerCtxKey.String()))
}

func (s *mdwMailerSuite) TestMailerPreviewWithDebugBuild() {
	s.mailer.AddPreview(&mailer.Mail{
		From:         "support@appy.org",
		To:           []string{"a@appy.org"},
		ReplyTo:      []string{"b@appy.org"},
		Cc:           []string{"c@appy.org"},
		Bcc:          []string{"d@appy.org"},
		Subject:      "mailers.user.welcome.subject",
		Template:     "mailers/user/welcome",
		TemplateData: H{},
	})

	s.mailer.AddPreview(&mailer.Mail{
		From:         "support@appy.org",
		To:           []string{"a@appy.org"},
		ReplyTo:      []string{"b@appy.org"},
		Cc:           []string{"c@appy.org"},
		Bcc:          []string{"d@appy.org"},
		Subject:      "mailers.user.welcome.subject",
		Template:     "mailers/user/error",
		TemplateData: H{},
	})

	mdwMailer(s.mailer, s.i18n, s.server)

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath, nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/missing&locale=en&ext=html", nil, nil)
		s.Equal(http.StatusNotFound, recorder.Code)
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/error&locale=en&ext=html", nil, nil)
		s.Equal(http.StatusInternalServerError, recorder.Code)
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/error&locale=en&ext=txt", nil, nil)
		s.Equal(http.StatusInternalServerError, recorder.Code)
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=en&ext=html", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "I&#39;m a mailer html version.")
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=zh-CN&ext=html", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "我是寄信者网页版。")
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=zh-TW&ext=html", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "我是寄信者網頁版。")
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=en&ext=txt", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.NotContains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "I'm a mailer txt version.")
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=zh-CN&ext=txt", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.NotContains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "我是寄信者文字版。")
	}

	{
		recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=zh-TW&ext=txt", nil, nil)
		s.Equal(http.StatusOK, recorder.Code)
		s.NotContains(recorder.Body.String(), "<!DOCTYPE html>")
		s.Contains(recorder.Body.String(), "我是寄信者文字版。")
	}
}

func (s *mdwMailerSuite) TestMailerPreviewWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() { support.Build = support.DebugBuild }()

	mdwMailer(s.mailer, s.i18n, s.server)

	recorder := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath, nil, nil)
	s.Equal(http.StatusNotFound, recorder.Code)

	recorder = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?name=mailers/user/welcome&locale=en&ext=html", nil, nil)
	s.Equal(http.StatusNotFound, recorder.Code)
}

func TestMdwMailerSuite(t *testing.T) {
	test.Run(t, new(mdwMailerSuite))
}
