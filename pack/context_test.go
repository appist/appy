package pack

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/appist/appy/view"
)

type contextSuite struct {
	test.Suite
	asset      *support.Asset
	config     *support.Config
	i18n       *support.I18n
	logger     *support.Logger
	server     *Server
	viewEngine *view.Engine
}

func (s *contextSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "testdata/context")
	s.config = support.NewConfig(s.asset, s.logger)
	s.i18n = support.NewI18n(s.asset, s.config, s.logger)
	s.viewEngine = view.NewEngine(s.asset, s.config, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger)
}

func (s *contextSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *contextSuite) TestCSRFTemplateField() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "foobar")

	{
		s.Equal(`<input type="hidden" name="authenticity_token" value="foobar">`, c.CSRFAuthenticityTemplateField())
	}

	{
		c.Set(mdwCSRFAuthenticityFieldNameCtxKey.String(), "x_authenticity_token")
		s.Equal(`<input type="hidden" name="x_authenticity_token" value="foobar">`, c.CSRFAuthenticityTemplateField())
	}
}

func (s *contextSuite) TestCSRFToken() {
	c, _ := NewTestContext(httptest.NewRecorder())
	s.Equal("", c.CSRFAuthenticityToken())

	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "foobar")
	s.Equal("foobar", c.CSRFAuthenticityToken())
}

func (s *contextSuite) TestI18n() {
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(mdwI18nCtxKey.String(), s.i18n)

	s.Equal([]string{"en", "zh-CN", "zh-TW"}, c.Locales())
	s.Equal("en", c.Locale())
	s.Equal("Test", c.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", c.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", c.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", c.T("body.message", 2, H{"Name": "tester"}))

	c.SetLocale("zh-CN")
	s.Equal("zh-CN", c.Locale())
	s.Equal("测试", c.T("title.test"))
	s.Equal("嗨, tester! 您有0则讯息。", c.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("嗨, tester! 您有1则讯息。", c.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("嗨, tester! 您有2则讯息。", c.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", c.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", c.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", c.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", c.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))
}

func (s *contextSuite) TestDeliver() {
	s.config.AppyEnv = "test"
	m := mailer.NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(mdwI18nCtxKey.String(), s.i18n)
	c.Set(mdwMailerCtxKey.String(), m)

	mail := &mailer.Mail{
		From:     "support@appist.io",
		To:       []string{"jane@appist.io"},
		ReplyTo:  []string{"john@appist.io", "mary@appist.io"},
		Cc:       []string{"elaine@appist.io", "kerry@appist.io"},
		Bcc:      []string{"joel@appist.io", "daniel@appist.io"},
		Subject:  "mailers.user.verifyAccount.subject",
		Template: "mailers/user/verify_account",
		TemplateData: support.H{
			"username": "cayter",
		},
	}

	err := c.Deliver(mail)
	s.Nil(err)

	deliveries := m.Deliveries()
	s.Equal(1, len(deliveries))
	s.Contains(deliveries[0].HTML, "cayter")
	s.Contains(deliveries[0].HTML, "测试")
	s.Contains(deliveries[0].HTML, "Hi, John Doe! You have 2 messages.")
	s.Contains(deliveries[0].Text, "cayter")
	s.Contains(deliveries[0].Text, "测试")
	s.Contains(deliveries[0].Text, "Hi, John Doe! You have 2 messages.")
}

func (s *contextSuite) TestHTML() {
	server := NewServer(s.asset, s.config, s.logger)
	server.Use(mdwLogger(s.logger))
	server.Use(mdwI18n(s.i18n))
	server.Use(mdwViewEngine(s.asset, s.config, s.logger, nil))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "mailers/user/welcome.html", H{})
	})

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "I&#39;m a mailer html version.")
	}

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept-Language", "zh-TW")
		server.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "我是寄信者網頁版。")
	}

	{
		server.GET("/error", func(c *Context) {
			c.HTML(http.StatusOK, "mailers/user/error.html", H{})
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		server.ServeHTTP(w, req)

		s.Equal(http.StatusInternalServerError, w.Code)
	}
}

func (s *contextSuite) TestHTMLMissingTemplate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	server := NewServer(s.asset, s.config, s.logger)
	server.Use(mdwLogger(s.logger))
	server.Use(mdwI18n(s.i18n))
	server.Use(mdwViewEngine(s.asset, s.config, s.logger, nil))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "dummy/index.html", H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *contextSuite) TestViewEngineWithDebugBuild() {
	server := NewServer(s.asset, s.config, s.logger)
	server.Use(mdwLogger(s.logger))
	server.Use(mdwI18n(s.i18n))
	server.Use(mdwViewEngine(s.asset, s.config, s.logger, map[string]interface{}{
		"add": func(a, b int) int {
			return a + b
		},
	}))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "home/index.html", H{})
	})

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "This content will be yielded in the layout above. 3")
		s.Contains(w.Body.String(), `b("ws://:`+LiveReloadWSPort+LiveReloadPath+`")`)
	}

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.TLS = &tls.ConnectionState{}
		server.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "This content will be yielded in the layout above. 3")
		s.Contains(w.Body.String(), `b("wss://:`+LiveReloadWSSPort+LiveReloadPath+`")`)
	}
}

func (s *contextSuite) TestViewEngineWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	s.asset = support.NewAsset(http.Dir("testdata/context"), "")
	s.viewEngine = view.NewEngine(s.asset, s.config, s.logger)

	server := NewServer(s.asset, s.config, s.logger)
	server.Use(mdwLogger(s.logger))
	server.Use(mdwI18n(s.i18n))
	server.Use(mdwViewEngine(s.asset, s.config, s.logger, map[string]interface{}{
		"add": func(a, b int) int {
			return a + b
		},
	}))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "home/index.html", H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "This content will be yielded in the layout above.")
	s.NotContains(w.Body.String(), `b("ws://:`+LiveReloadWSPort+LiveReloadPath+`")`)
}

func TestContextSuite(t *testing.T) {
	test.Run(t, new(contextSuite))
}
