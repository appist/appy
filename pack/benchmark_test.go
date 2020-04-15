package pack

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
)

func newServer() *Server {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	asset := support.NewAsset(http.Dir("testdata/context"), "testdata/context")
	logger, _, _ := support.NewTestLogger()
	config := support.NewConfig(asset, logger)
	i18n := support.NewI18n(asset, config, logger)
	ml := mailer.NewEngine(asset, config, i18n, logger, nil)

	server := NewServer(asset, config, logger)
	server.Use(mdwLogger(logger))
	server.Use(mdwI18n(i18n))
	server.Use(mdwMailer(ml, i18n, server))
	server.Use(mdwViewEngine(asset, config, logger, nil))
	server.Use(mdwRealIP())
	server.Use(mdwReqID())
	server.Use(mdwReqLogger(config, logger))
	server.Use(mdwGzip(config))
	server.Use(mdwHealthCheck(config.HTTPHealthCheckPath, server))
	server.Use(mdwPrerender(config, logger))
	server.Use(mdwCSRF(config, logger))
	server.Use(mdwSecure(config))
	server.Use(mdwAPIOnly())
	server.Use(mdwSession(config))
	server.Use(mdwRecovery(logger))

	return server
}

func testRequest(B *testing.B, server *Server, method, path string) {
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
