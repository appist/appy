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

type ContextSuite struct {
	test.Suite
	assets     *support.Assets
	config     *support.Config
	i18n       *support.I18n
	logger     *support.Logger
	mailer     *am.Mailer
	viewEngine *support.ViewEngine
}

func (s *ContextSuite) SetupTest() {
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
	s.mailer = am.NewMailer(s.config, s.i18n, s.viewEngine)
}

func (s *ContextSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ContextSuite) TestI18n() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(i18nCtxKey.String(), s.i18n)
	s.Equal([]string{"en", "zh-CN", "zh-TW"}, c.Locales())

	s.Equal("en", c.Locale())
	s.Equal("Test", c.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", c.T("body.message", 0, support.H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", c.T("body.message", 1, support.H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", c.T("body.message", 2, support.H{"Name": "tester"}))

	c.SetLocale("zh-CN")
	s.Equal("zh-CN", c.Locale())
	s.Equal("测试", c.T("title.test"))
	s.Equal("嗨, tester! 您有0则讯息。", c.T("body.message", 0, support.H{"Name": "tester"}))
	s.Equal("嗨, tester! 您有1则讯息。", c.T("body.message", 1, support.H{"Name": "tester"}))
	s.Equal("嗨, tester! 您有2则讯息。", c.T("body.message", 2, support.H{"Name": "tester"}))

	s.Equal("測試", c.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", c.T("body.message", 0, support.H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", c.T("body.message", 1, support.H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", c.T("body.message", 2, support.H{"Name": "tester"}, "zh-TW"))
}

func (s *ContextSuite) TestRenderHTML() {
	server := NewServer(s.assets, s.config, s.logger)
	server.Use(I18n(s.i18n))
	server.Use(ViewEngine(s.viewEngine))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "mailers/user/welcome.html", support.H{})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "I&#39;m a mailer html version.")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "zh-TW")
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "我是寄信者網頁版。")

	server.GET("/error", func(c *Context) {
		c.HTML(http.StatusOK, "mailers/user/error.html", support.H{})
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/error", nil)
	server.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ContextSuite) TestRenderHTMLMissingTemplate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	server := NewServer(s.assets, s.config, s.logger)
	server.Use(I18n(s.i18n))
	server.Use(ViewEngine(s.viewEngine))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "dummy/index.html", support.H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ContextSuite) TestSendMail() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(i18nCtxKey.String(), s.i18n)
	c.Set(mailerCtxKey.String(), s.mailer)

	err := c.SendEmail(am.Email{
		From:     "foo@appist.io",
		To:       []string{"bar@appist.io"},
		ReplyTo:  []string{},
		Bcc:      []string{},
		Cc:       []string{},
		Sender:   "foo@appist.io",
		Subject:  "Welcome",
		Template: "mailers/user/welcome",
	})

	s.EqualError(err, "dial tcp: missing address")
}

func (s *ContextSuite) TestSendMailWithTLS() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(i18nCtxKey.String(), s.i18n)
	c.Set(mailerCtxKey.String(), s.mailer)

	err := c.SendEmailWithTLS(am.Email{
		From:     "foo@appist.io",
		To:       []string{"bar@appist.io"},
		ReplyTo:  []string{},
		Bcc:      []string{},
		Cc:       []string{},
		Sender:   "foo@appist.io",
		Subject:  "Welcome",
		Template: "mailers/user/welcome",
	})

	s.EqualError(err, "dial tcp: missing address")
}

func (s *ContextSuite) TestViewEngineWithDebugBuild() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	server := NewServer(s.assets, s.config, s.logger)
	server.Use(I18n(s.i18n))
	server.Use(ViewEngine(s.viewEngine))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "home/index.html", support.H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "This content will be yielded in the layout above.")
}

func (s *ContextSuite) TestViewEngineWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	s.assets = support.NewAssets(nil, "", http.Dir("../support/testdata"))
	s.viewEngine = support.NewViewEngine(s.assets)

	server := NewServer(s.assets, s.config, s.logger)
	server.Use(I18n(s.i18n))
	server.Use(ViewEngine(s.viewEngine))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "home/index.html", support.H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "This content will be yielded in the layout above.")
}

func TestContextSuite(t *testing.T) {
	test.RunSuite(t, new(ContextSuite))
}
