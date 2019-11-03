package appy

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type RecoverySuite struct {
	TestSuite
	config   *Config
	logger   *Logger
	recorder *httptest.ResponseRecorder
	server   *Server
}

func (s *RecoverySuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.logger = NewLogger(DebugBuild)
	s.config = NewConfig(DebugBuild, s.logger, nil)
	s.recorder = httptest.NewRecorder()
	s.server = NewServer(s.config, s.logger, nil, nil)
}

func (s *RecoverySuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *RecoverySuite) TestPanicRenders500() {
	s.server.router.Use(SessionManager(s.config))
	s.server.router.GET("/test", func(ctx *Context) {
		session := DefaultSession(ctx)
		session.Set("username", "dummy")
		panic(errors.New("error"))
	})

	request, _ := http.NewRequest("GET", "/test?age=10", nil)
	request.Header.Set("X-Testing", "1")
	s.server.router.ServeHTTP(s.recorder, request)

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>500 Internal Server Error</title>")
	s.Contains(s.recorder.Body.String(), "username: dummy")
	s.Contains(s.recorder.Body.String(), "X-Testing: 1")
	s.Contains(s.recorder.Body.String(), "age: 10")
}

func (s *RecoverySuite) TestBrokenPipeErrorHandling() {
	s.server.router.GET("/test", func(ctx *Context) {
		panic(&net.OpError{Err: &os.SyscallError{Err: errors.New("broken pipe")}})
	})

	request, _ := http.NewRequest("GET", "/test", nil)
	s.server.router.ServeHTTP(s.recorder, request)

	s.Contains(s.recorder.Body.String(), "broken pipe")
}

func TestRecovery(t *testing.T) {
	RunTestSuite(t, new(RecoverySuite))
}
