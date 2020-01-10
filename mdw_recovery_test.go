package appy

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type RecoverySuite struct {
	TestSuite
	asset    *Asset
	config   *Config
	logger   *Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
	server   *Server
	support  Supporter
}

func (s *RecoverySuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata"), nil)
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.recorder = httptest.NewRecorder()
	s.server = NewServer(s.asset, s.config, s.logger, s.support)
}

func (s *RecoverySuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *RecoverySuite) TestPanicRenders500WithDebug() {
	s.server.Use(SessionManager(s.config))
	s.server.Use(Recovery(s.logger))
	s.server.GET("/test", func(c *Context) {
		session := c.Session()
		session.Set("username", "dummy")
		panic(errors.New("error"))
	})

	req, _ := http.NewRequest("GET", "/test?age=10", nil)
	req.Header.Set("X-Testing", "1")
	s.server.router.ServeHTTP(s.recorder, req)

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>500 Internal Server Error</title>")
	s.Contains(s.recorder.Body.String(), "username: dummy")
	s.Contains(s.recorder.Body.String(), "X-Testing: 1")
	s.Contains(s.recorder.Body.String(), "age: 10")
}

func (s *RecoverySuite) TestPanicRenders500WithRelease() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	s.server = NewServer(s.asset, s.config, s.logger, s.support)
	s.server.Use(SessionManager(s.config))
	s.server.Use(Recovery(s.logger))
	s.server.GET("/test", func(c *Context) {
		session := c.Session()
		session.Set("username", "dummy")
		panic(errors.New("error"))
	})

	req, _ := http.NewRequest("GET", "/test?age=10", nil)
	req.Header.Set("X-Testing", "1")
	s.server.router.ServeHTTP(s.recorder, req)

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), `<p class="card-text">If you are the administrator of this website, then please read this web application's log file and/or the web server's log file to find out what went wrong.</p>`)
}

func (s *RecoverySuite) TestBrokenPipeErrorHandling() {
	s.server.Use(Recovery(s.logger))
	s.server.GET("/test", func(c *Context) {
		panic(&net.OpError{Err: &os.SyscallError{Err: errors.New("broken pipe")}})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	s.server.ServeHTTP(s.recorder, req)

	s.Contains(s.recorder.Body.String(), "broken pipe")
}

func TestRecoverySuite(t *testing.T) {
	RunTestSuite(t, new(RecoverySuite))
}
