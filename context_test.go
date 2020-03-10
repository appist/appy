package appy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type ContextSuite struct {
	TestSuite
	asset      *Asset
	config     *Config
	i18n       *I18n
	logger     *Logger
	server     *Server
	support    Supporter
	viewEngine *ViewEngine
}

func (s *ContextSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, _, _ = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata/context"), map[string]string{
		"docker": "testdata/context/.docker",
		"config": "testdata/context/configs",
		"locale": "testdata/context/pkg/locales",
		"view":   "testdata/context/pkg/views",
		"web":    "testdata/context/web",
	}, "")
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.i18n = NewI18n(s.asset, s.config, s.logger)
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger, s.support)
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

func (s *ContextSuite) TestDeliverMail() {
	s.config.AppyEnv = "test"
	mailer := NewMailer(s.asset, s.config, s.i18n, s.logger, s.server, nil)
	c, _ := NewTestContext(httptest.NewRecorder())
	c.Set(i18nCtxKey.String(), s.i18n)
	c.Set(mailerCtxKey.String(), mailer)

	mail := Mail{
		From:    "support@appist.io",
		To:      []string{"jane@appist.io"},
		ReplyTo: []string{"john@appist.io", "mary@appist.io"},
		Cc:      []string{"elaine@appist.io", "kerry@appist.io"},
		Bcc:     []string{"joel@appist.io", "daniel@appist.io"},
	}
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = H{
		"username": "cayter",
	}
	c.DeliverMail(mail)
	s.Equal(1, len(mailer.Deliveries()))
}

func (s *ContextSuite) TestHTML() {
	server := NewServer(s.asset, s.config, s.logger, s.support)
	server.Use(AttachLogger(s.logger))
	server.Use(AttachI18n(s.i18n))
	server.Use(AttachViewEngine(s.asset, s.config, s.logger, nil))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "mailers/user/welcome.html", H{})
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
		c.HTML(http.StatusOK, "mailers/user/error.html", H{})
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/error", nil)
	server.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ContextSuite) TestHTMLMissingTemplate() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	server := NewServer(s.asset, s.config, s.logger, s.support)
	server.Use(AttachLogger(s.logger))
	server.Use(AttachI18n(s.i18n))
	server.Use(AttachViewEngine(s.asset, s.config, s.logger, nil))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "dummy/index.html", H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *ContextSuite) TestViewEngineWithDebugBuild() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	server := NewServer(s.asset, s.config, s.logger, s.support)
	server.Use(AttachLogger(s.logger))
	server.Use(AttachI18n(s.i18n))
	server.Use(AttachViewEngine(s.asset, s.config, s.logger, map[string]interface{}{
		"add": func(a, b int) int {
			return a + b
		},
	}))
	server.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "home/index.html", H{})
	})
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "This content will be yielded in the layout above. 3")
	s.Contains(w.Body.String(), `b("ws://:`+LiveReloadWSPort+LiveReloadPath+`")`)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{}
	server.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "This content will be yielded in the layout above. 3")
	s.Contains(w.Body.String(), `b("wss://:`+LiveReloadWSSPort+LiveReloadPath+`")`)
}

func (s *ContextSuite) TestViewEngineWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	s.asset = NewAsset(http.Dir("testdata/context"), nil, "")
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)

	server := NewServer(s.asset, s.config, s.logger, s.support)
	server.Use(AttachLogger(s.logger))
	server.Use(AttachI18n(s.i18n))
	server.Use(AttachViewEngine(s.asset, s.config, s.logger, map[string]interface{}{
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
	RunTestSuite(t, new(ContextSuite))
}
