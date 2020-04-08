package pack

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwSecureSuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	logger *support.Logger
	buffer *bytes.Buffer
	writer *bufio.Writer
	server *Server
}

func (s *mdwSecureSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger)
}

func (s *mdwSecureSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwSecureSuite) TestSSL() {
	s.config.HTTPSSLRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "https://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *mdwSecureSuite) TestSSLRedirectDisabled() {
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "https://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *mdwSecureSuite) TestSSLRedirectEnabled() {
	s.config.HTTPSSLRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://www.example.com/foo", w.Header().Get("Location"))
}

func (s *mdwSecureSuite) TestSSLTemporaryRedirectEnabled() {
	s.config.HTTPSSLRedirect = true
	s.config.HTTPSSLTemporaryRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusFound, w.Code)
	s.Equal("https://www.example.com/foo", w.Header().Get("Location"))
}

func (s *mdwSecureSuite) TestNoAllowedHosts() {
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *mdwSecureSuite) TestGoodSingleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com"}
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *mdwSecureSuite) TestBadSingleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"sub.example.com"}
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *mdwSecureSuite) TestGoodMultipleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com", "sub.example.com"}
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://sub.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("bar", w.Body.String())
}

func (s *mdwSecureSuite) TestBadMultipleAllowedHosts() {
	s.config.HTTPAllowedHosts = []string{"www.example.com", "sub.example.com"}
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www3.example.com/foo", nil, nil)
	s.Equal(http.StatusForbidden, w.Code)
}

func (s *mdwSecureSuite) TestBasicSSLWithHost() {
	s.config.HTTPSSLHost = "secure.example.com"
	s.config.HTTPSSLRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
	s.Equal("https://secure.example.com/foo", w.Header().Get("Location"))
}

func (s *mdwSecureSuite) TestProxySSLWithHeaderOption() {
	s.config.HTTPSSLProxyHeaders = map[string]string{"X-Arbitrary-Header": "arbitrary-value"}
	s.config.HTTPSSLRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", H{"X-Arbitrary-Header": "arbitrary-value"}, nil)
	s.Equal(http.StatusOK, w.Code)
}

func (s *mdwSecureSuite) TestProxySSLWithWrongHeaderValue() {
	s.config.HTTPSSLProxyHeaders = map[string]string{"X-Arbitrary-Header": "arbitrary-value"}
	s.config.HTTPSSLRedirect = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", H{"X-Arbitrary-Header": "wrong-value"}, nil)
	s.Equal(http.StatusMovedPermanently, w.Code)
}

func (s *mdwSecureSuite) TestSTSHeader() {
	s.config.HTTPSTSSeconds = 315360000
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=315360000", w.Header().Get("Strict-Transport-Security"))
}

func (s *mdwSecureSuite) TestSTSHeaderWithSubdomain() {
	s.config.HTTPSTSSeconds = 315360000
	s.config.HTTPSTSIncludeSubdomains = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("max-age=315360000; includeSubdomains", w.Header().Get("Strict-Transport-Security"))
}

func (s *mdwSecureSuite) TestFrameDeny() {
	s.config.HTTPFrameDeny = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("DENY", w.Header().Get("X-Frame-Options"))
}

func (s *mdwSecureSuite) TestCustomFrameValue() {
	s.config.HTTPCustomFrameOptionsValue = "SAMEORIGIN"
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func (s *mdwSecureSuite) TestCustomFrameValueWithDeny() {
	s.config.HTTPFrameDeny = true
	s.config.HTTPCustomFrameOptionsValue = "SAMEORIGIN"
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func (s *mdwSecureSuite) TestContentTypeNosniff() {
	s.config.HTTPContentTypeNosniff = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("nosniff", w.Header().Get("X-Content-Type-Options"))
}

func (s *mdwSecureSuite) TestXSSProtection() {
	s.config.HTTPBrowserXSSFilter = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func (s *mdwSecureSuite) TestReferrerPolicy() {
	s.config.HTTPReferrerPolicy = "strict-origin-when-cross-origin"
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func (s *mdwSecureSuite) TestContentSecurityPolicy() {
	s.config.HTTPContentSecurityPolicy = "default-src 'self'"
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("default-src 'self'", w.Header().Get("Content-Security-Policy"))
}

func (s *mdwSecureSuite) TestIENoOpen() {
	s.config.HTTPIENoOpen = true
	s.server.Use(mdwSecure(s.config))
	s.server.GET("/foo", func(c *Context) {
		c.String(http.StatusOK, "bar")
	})

	w := s.server.TestHTTPRequest("GET", "http://www.example.com/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("noopen", w.Header().Get("X-Download-Options"))
}

func TestMdwSecureSuite(t *testing.T) {
	test.Run(t, new(mdwSecureSuite))
}
