package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type ServerSuite struct {
	test.Suite
	assetsMngr *support.AssetsMngr
	config     *support.Config
	logger     *support.Logger
	buffer     *bytes.Buffer
	writer     *bufio.Writer
}

func (s *ServerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	layout := map[string]string{
		"docker": "../support/testdata/.docker",
		"config": "../support/testdata/pkg/config",
		"locale": "../support/testdata/pkg/locales",
		"view":   "../support/testdata/pkg/views",
		"web":    "../support/testdata/web",
	}
	s.assetsMngr = support.NewAssetsMngr(layout, "", nil)
	s.logger, s.buffer, s.writer = support.NewFakeLogger()
	s.config = support.NewConfig(s.assetsMngr, s.logger)
}

func (s *ServerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ServerSuite) TestNewServerWithSSLEnabled() {
	s.config.HTTPSSLEnabled = true
	server := NewServer(s.assetsMngr, s.config, s.logger)
	s.NotNil(server.Config())
	s.NotNil(server.GRPC())
	s.NotNil(server.HTTP())
	s.NotNil(server.HTTPS())
	s.NotNil(server.router)
	s.Equal(0, len(server.Middleware()))
	s.Equal("localhost:3000", server.HTTP().Addr)
	s.Equal("localhost:3443", server.HTTPS().Addr)
}

func (s *ServerSuite) TestNewServerWithoutSSLEnabled() {
	server := NewServer(s.assetsMngr, s.config, s.logger)
	s.NotNil(server.Config())
	s.NotNil(server.GRPC())
	s.NotNil(server.HTTP())
	s.NotNil(server.HTTPS())
	s.NotNil(server.router)
	s.Equal(0, len(server.Middleware()))
	s.Equal("localhost:3000", server.HTTP().Addr)
	s.Equal("localhost:3443", server.HTTPS().Addr)
}

func (s *ServerSuite) TestIsSSLCertsExisted() {
	server := NewServer(s.assetsMngr, s.config, s.logger)
	s.Equal(false, server.IsSSLCertExisted())

	s.config.HTTPSSLCertPath = "testdata/ssl"
	server = NewServer(s.assetsMngr, s.config, s.logger)
	s.Equal(true, server.IsSSLCertExisted())
}

func (s *ServerSuite) TestInfo() {
	server := NewServer(s.assetsMngr, s.config, s.logger)
	output := server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000")

	s.config.HTTPSSLEnabled = true
	server = NewServer(s.assetsMngr, s.config, s.logger)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000, https://localhost:3443")

	s.config.HTTPHost = "0.0.0.0"
	server = NewServer(s.assetsMngr, s.config, s.logger)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000, https://0.0.0.0:3443")
}

func (s *ServerSuite) TestRouting() {
	server := NewServer(s.assetsMngr, s.config, s.logger)
	s.Equal(server.BasePath(), "/")

	server.Any("/foo", func(c *Context) { c.String(http.StatusOK, "bar") })
	methods := []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST", "TRACE"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/foo", nil, nil)
		s.Equal(http.StatusOK, w.Code)
		s.Equal("bar", w.Body.String())
	}

	server.Handle("CONNECT", "/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.DELETE("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.GET("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.HEAD("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.OPTIONS("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.PATCH("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.PUT("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	server.POST("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/bar", nil, nil)
		s.Equal(http.StatusOK, w.Code)
		s.Equal("foo", w.Body.String())
	}

	count := 1
	v1 := server.Group("/v1")
	v1.Use(func(c *Context) {
		count = 10
		c.Next()
	})
	v1.Any("/foo", func(c *Context) { c.String(http.StatusOK, "bar") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST", "TRACE"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/v1/foo", nil, nil)
		s.Equal(http.StatusOK, w.Code)
		s.Equal("bar", w.Body.String())
	}
	s.Equal(10, count)

	v1.Handle("CONNECT", "/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.DELETE("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.GET("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.HEAD("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.OPTIONS("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.PATCH("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.PUT("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	v1.POST("/bar", func(c *Context) { c.String(http.StatusOK, "foo") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/v1/bar", nil, nil)
		s.Equal(http.StatusOK, w.Code)
		s.Equal("foo", w.Body.String())
	}

	routes := server.Routes()
	s.Equal(35, len(routes))
}

func TestServerSuite(t *testing.T) {
	test.RunSuite(t, new(ServerSuite))
}
