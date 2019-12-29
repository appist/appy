package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	am "github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type MailerSuite struct {
	test.Suite
	assets     *support.Assets
	config     *support.Config
	i18n       *support.I18n
	logger     *support.Logger
	mailer     *am.Mailer
	server     *Server
	viewEngine *support.ViewEngine
}

func (s *MailerSuite) SetupTest() {
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
	s.viewEngine = support.NewViewEngine(s.assets)
	s.server = NewServer(s.assets, s.config, s.logger)
	s.mailer = am.NewMailer(s.config, s.i18n, s.viewEngine)
}

func (s *MailerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *MailerSuite) TestMailerWithDebugBuild() {
	c, _ := NewTestContext(httptest.NewRecorder())
	Mailer(s.i18n, s.mailer, s.server)(c)
	mailer, _ := c.Get(mailerCtxKey.String())
	s.NotNil(mailer)

	w := s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL, nil, nil)
	s.Equal(http.StatusOK, w.Code)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)

	s.mailer.AddPreview(am.Email{
		From:     "foo@appist.io",
		To:       []string{"bar@appist.io"},
		ReplyTo:  []string{},
		Bcc:      []string{},
		Cc:       []string{},
		Sender:   "foo",
		Subject:  "Welcome",
		Template: "mailers/user/welcome",
	})

	s.mailer.AddPreview(am.Email{
		From:     "foo@appist.io",
		To:       []string{"bar@appist.io"},
		ReplyTo:  []string{},
		Bcc:      []string{},
		Cc:       []string{},
		Sender:   "foo",
		Subject:  "Error",
		Template: "mailers/user/error",
	})

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL, nil, nil)
	s.Equal(http.StatusOK, w.Code)

	s.Panics(func() {
		s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/error&ext=html", nil, nil)
	})

	s.Panics(func() {
		s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/error&ext=txt", nil, nil)
	})

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/welcome", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "I&#39;m a mailer html version.")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/welcome&ext=html", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "I&#39;m a mailer html version.")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/welcome&ext=txt", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "I&#39;m a mailer txt version.")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/welcome&ext=html&locale=zh-TW", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "我是寄信者網頁版。")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewBaseURL+"/preview?name=mailers/user/welcome&ext=txt&locale=zh-CN", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "我是寄信者文字版。")
}

func TestMailerSuite(t *testing.T) {
	test.RunSuite(t, new(MailerSuite))
}
