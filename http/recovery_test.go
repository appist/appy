package http

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type RecoverySuite struct {
	test.Suite
	assetsMngr *support.AssetsMngr
	config     *support.Config
	logger     *support.Logger
	buffer     *bytes.Buffer
	writer     *bufio.Writer
	recorder   *httptest.ResponseRecorder
	server     *Server
}

func (s *RecoverySuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = support.NewFakeLogger()
	s.assetsMngr = support.NewAssetsMngr(nil, "", http.Dir("../support/testdata"))
	s.config = support.NewConfig(s.assetsMngr, s.logger)
	s.recorder = httptest.NewRecorder()
	s.server = NewServer(s.assetsMngr, s.config, s.logger)
}

func (s *RecoverySuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *RecoverySuite) TestPanicRenders500() {
	s.server.Use(SessionMngr(s.config))
	s.server.Use(Recovery(s.logger))
	s.server.GET("/test", func(c *Context) {
		session := DefaultSession(c)
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

func (s *RecoverySuite) TestBrokenPipeErrorHandling() {
	s.server.Use(Recovery(s.logger))
	s.server.GET("/test", func(c *Context) {
		panic(&net.OpError{Err: &os.SyscallError{Err: errors.New("broken pipe")}})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	s.server.ServeHTTP(s.recorder, req)

	s.Contains(s.recorder.Body.String(), "broken pipe")
}

func TestRecovery(t *testing.T) {
	test.RunSuite(t, new(RecoverySuite))
}
