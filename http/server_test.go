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
	assets *support.Assets
	config *support.Config
	logger *support.Logger
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func (s *ServerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	layout := map[string]string{
		"docker": "../support/testdata/.docker",
		"config": "../support/testdata/configs",
		"locale": "../support/testdata/pkg/locales",
		"view":   "../support/testdata/pkg/views",
		"web":    "../support/testdata/web",
	}
	s.assets = support.NewAssets(layout, "", nil)
	s.logger, s.buffer, s.writer = support.NewFakeLogger()
	s.config = support.NewConfig(s.assets, s.logger)
}

func (s *ServerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ServerSuite) TestNewServerWithSSLEnabled() {
	s.config.HTTPSSLEnabled = true
	server := NewServer(s.assets, s.config, s.logger)
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
	server := NewServer(s.assets, s.config, s.logger)
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
	server := NewServer(s.assets, s.config, s.logger)
	s.Equal(false, server.IsSSLCertExisted())

	s.config.HTTPSSLCertPath = "testdata/ssl"
	server = NewServer(s.assets, s.config, s.logger)
	s.Equal(true, server.IsSSLCertExisted())
}

func (s *ServerSuite) TestInfo() {
	server := NewServer(s.assets, s.config, s.logger)
	output := server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000")

	s.config.HTTPSSLEnabled = true
	server = NewServer(s.assets, s.config, s.logger)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000, https://localhost:3443")

	s.config.HTTPHost = "0.0.0.0"
	server = NewServer(s.assets, s.config, s.logger)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: ../support/testdata/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000, https://0.0.0.0:3443")
}

func (s *ServerSuite) TestRouting() {
	server := NewServer(s.assets, s.config, s.logger)
	server.ServeNoRoute()
	s.Equal(server.BasePath(), "/")

	w := server.TestHTTPRequest("GET", "/foobar", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")

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

func (s *ServerSuite) TestCSRWithDebugBuild() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := NewServer(s.assets, s.config, s.logger)
	server.ServeSPA("/", nil)
	w := server.TestHTTPRequest("GET", "/", nil, nil)

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(502, w.Code)

	s.config.HTTPSSLEnabled = true
	server = NewServer(s.assets, s.config, s.logger)
	server.ServeSPA("/", nil)
	w = server.TestHTTPRequest("GET", "/", nil, nil)

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(502, w.Code)
}

func (s *ServerSuite) TestCSRWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		support.Build = support.DebugBuild
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := NewServer(s.assets, s.config, s.logger)
	server.ServeSPA("/", http.Dir("./testdata/csr"))
	w := server.TestHTTPRequest("GET", "/", nil, nil)

	s.Equal(200, w.Code)
	s.Contains(w.Body.String(), `<div id="app">we build apps</div>`)

	server = NewServer(s.assets, s.config, s.logger)
	server.GET("/ssr", func(c *Context) {
		c.String(http.StatusOK, "foobar")
	})
	server.ServeSPA("/", http.Dir("./testdata/csr"))
	server.ServeSPA("/ssr", http.Dir("./testdata/csr"))
	server.ServeSPA("/tools", http.Dir("./testdata/csr/tools"))
	server.ServeNoRoute()

	w = server.TestHTTPRequest("GET", "/ssr/foo", nil, nil)
	s.Equal(404, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")

	w = server.TestHTTPRequest("GET", "/foo", nil, nil)
	s.Equal(404, w.Code)
	s.Contains(w.Body.String(), "404 page not found")

	w = server.TestHTTPRequest("GET", "/tools", nil, nil)
	s.Equal(200, w.Code)
	s.Contains(w.Body.String(), `<div id="app">we build another SPA</div>`)

	w = server.TestHTTPRequest("GET", "/ssr", nil, nil)
	s.Equal(200, w.Code)
	s.Contains(w.Body.String(), "foobar")

	w = server.TestHTTPRequest("GET", "/.ssr", nil, nil)
	s.Equal(404, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")

	server = NewServer(s.assets, s.config, s.logger)
	server.ServeSPA("/", nil)
	server.ServeNoRoute()

	w = server.TestHTTPRequest("GET", "/", nil, nil)
	s.Equal(404, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")
}

func TestServerSuite(t *testing.T) {
	test.RunSuite(t, new(ServerSuite))
}
