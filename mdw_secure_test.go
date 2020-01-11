package appy

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"testing"
)

type SecureSuite struct {
	TestSuite
	asset   *Asset
	config  *Config
	logger  *Logger
	buffer  *bytes.Buffer
	writer  *bufio.Writer
	server  *Server
	support Supporter
}

func (s *SecureSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata"), nil)
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.server = NewServer(s.asset, s.config, s.logger, s.support)
}

func (s *SecureSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *SecureSuite) TestSSL() {
	s.config.HTTPSSLRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "https://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *SecureSuite) TestSSLRedirectDisabled() {
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "https://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *SecureSuite) TestSSLRedirectEnabled() {
	s.config.HTTPSSLRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://www.example.com/foo", w.Header().Get("Location"))
}

func (s *SecureSuite) TestSSLTemporaryRedirectEnabled() {
	s.config.HTTPSSLRedirect = true
	s.config.HTTPSSLTemporaryRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusFound, w.Code)
	s.Equal("https://www.example.com/foo", w.Header().Get("Location"))
}

func (s *SecureSuite) TestNoAllowedHosts() {
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *SecureSuite) TestGoodSingleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com"}
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *SecureSuite) TestBadSingleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"sub.example.com"}
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *SecureSuite) TestGoodMultipleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com", "sub.example.com"}
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://sub.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *SecureSuite) TestBadMultipleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com", "sub.example.com"}
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www3.example.com/foo", nil, nil)
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *SecureSuite) TestBasicSSLWithHost() {
	s.config.HTTPSSLHost = "secure.example.com"
	s.config.HTTPSSLRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://secure.example.com/foo", w.Header().Get("Location"))
}

func (s *SecureSuite) TestProxySSLWithHeaderOption() {
	s.config.HTTPSSLProxyHeaders = map[string]string{"X-Arbitrary-Header": "arbitrary-value"}
	s.config.HTTPSSLRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", H{"X-Arbitrary-Header": "arbitrary-value"}, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *SecureSuite) TestProxySSLWithWrongHeaderValue() {
	s.config.HTTPSSLProxyHeaders = map[string]string{"X-Arbitrary-Header": "arbitrary-value"}
	s.config.HTTPSSLRedirect = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", H{"X-Arbitrary-Header": "wrong-value"}, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
}

func (s *SecureSuite) TestSTSHeader() {
	s.config.HTTPSTSSeconds = 315360000
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=315360000", w.Header().Get("Strict-Transport-Security"))
}

func (s *SecureSuite) TestSTSHeaderWithSubdomain() {
	s.config.HTTPSTSSeconds = 315360000
	s.config.HTTPSTSIncludeSubdomains = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=315360000; includeSubdomains", w.Header().Get("Strict-Transport-Security"))
}

func (s *SecureSuite) TestFrameDeny() {
	s.config.HTTPFrameDeny = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("DENY", w.Header().Get("X-Frame-Options"))
}

func (s *SecureSuite) TestCustomFrameValue() {
	s.config.HTTPCustomFrameOptionsValue = "SAMEORIGIN"
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func (s *SecureSuite) TestCustomFrameValueWithDeny() {
	s.config.HTTPFrameDeny = true
	s.config.HTTPCustomFrameOptionsValue = "SAMEORIGIN"
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func (s *SecureSuite) TestContentTypeNosniff() {
	s.config.HTTPContentTypeNosniff = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("nosniff", w.Header().Get("X-Content-Type-Options"))
}

func (s *SecureSuite) TestXSSProtection() {
	s.config.HTTPBrowserXSSFilter = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func (s *SecureSuite) TestReferrerPolicy() {
	s.config.HTTPReferrerPolicy = "strict-origin-when-cross-origin"
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func (s *SecureSuite) TestContentSecurityPolicy() {
	s.config.HTTPContentSecurityPolicy = "default-src 'self'"
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("default-src 'self'", w.Header().Get("Content-Security-Policy"))
}

func (s *SecureSuite) TestIENoOpen() {
	s.config.HTTPIENoOpen = true
	s.server.Use(Secure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("noopen", w.Header().Get("X-Download-Options"))
}

func TestSecureSuite(t *testing.T) {
	RunTestSuite(t, new(SecureSuite))
}
