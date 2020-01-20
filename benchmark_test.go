//+build benchmark

package appy

import (
	"io"
	"net/http"
	"os"
	"testing"
)

func newServer() *Server {
	asset := NewAsset(http.Dir("testdata/app"), map[string]string{
		"docker": "testdata/app/.docker",
		"config": "testdata/app/configs",
		"locale": "testdata/app/pkg/locales",
		"view":   "testdata/app/pkg/views",
		"web":    "testdata/config/web",
	})
	support := &Support{}
	logger, _, _ := NewFakeLogger()
	config := NewConfig(asset, logger, support)
	i18n := NewI18n(asset, config, logger)
	server := NewServer(asset, config, logger, support)
	mailer := NewMailer(asset, config, i18n, logger, server, nil)

	server.Use(AttachLogger(logger))
	server.Use(AttachI18n(i18n))
	server.Use(AttachMailer(mailer))
	server.Use(AttachViewEngine(asset, config, logger, nil))
	server.Use(RealIP())
	server.Use(RequestID())
	server.Use(RequestLogger(config, logger))
	server.Use(Gzip(config))
	server.Use(HealthCheck(config.HTTPHealthCheckURL))
	server.Use(Prerender(config, logger))
	server.Use(CSRF(config, logger, support))
	server.Use(Secure(config))
	server.Use(APIOnlyResponse())
	server.Use(SessionManager(config))
	server.Use(Recovery(logger))

	return server
}

func testRequest(B *testing.B, server *Server, method, path string) {
	B.ReportAllocs()
	B.ResetTimer()

	for i := 0; i < B.N; i++ {
		server.TestHTTPRequest("GET", "/ping", nil, nil)
	}
}

func BenchmarkServerParam(B *testing.B) {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := newServer()
	server.GET("/user/:name", func(c *Context) {})
	testRequest(B, server, "GET", "/user/gordon")
}

func BenchmarkServerParam5(B *testing.B) {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := newServer()
	server.GET("/:a/:b/:c/:d/:e", func(c *Context) {})
	testRequest(B, server, "GET", "/test/test/test/test/test")
}

func BenchmarkServerParam20(B *testing.B) {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := newServer()
	server.GET("/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t", func(c *Context) {})
	testRequest(B, server, "GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t")
}

func BenchmarkServerParamWrite(B *testing.B) {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := newServer()
	server.GET("/user/:name", func(c *Context) {
		io.WriteString(c.Writer, c.Params.ByName("name"))
	})
	testRequest(B, server, "GET", "/user/gordon")
}
