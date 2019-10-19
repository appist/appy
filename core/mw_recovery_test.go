package core

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type RecoverySuite struct {
	test.Suite
	config   AppConfig
	logger   *AppLogger
	recorder *httptest.ResponseRecorder
	server   AppServer
}

func (s *RecoverySuite) SetupTest() {
	s.logger, _ = newLogger(newLoggerConfig())
	s.config, _ = newConfig(nil, s.logger)
	s.config.HTTPSessionSecrets = [][]byte{[]byte("481e5d98a31585148b8b1dfb6a3c0465")}
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.recorder = httptest.NewRecorder()
	s.server = newServer(nil, s.config, s.logger, nil)
}

func (s *RecoverySuite) TearDownTest() {
}

func (s *RecoverySuite) TestPanicRenders500() {
	s.server.Router.Use(SessionManager(s.config))
	s.server.Router.GET("/test", func(ctx *Context) {
		session := DefaultSession(ctx)
		session.Set("username", "dummy")
		panic(errors.New("error"))
	})

	request, _ := http.NewRequest("GET", "/test?age=10", nil)
	request.Header.Set("X-Testing", "1")
	s.server.Router.ServeHTTP(s.recorder, request)

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>500 Internal Server Error</title>")
	s.Contains(s.recorder.Body.String(), "username: dummy")
	s.Contains(s.recorder.Body.String(), "X-Testing: 1")
	s.Contains(s.recorder.Body.String(), "age: 10")
}

func (s *RecoverySuite) TestBrokenPipeErrorHandling() {
	s.server.Router.GET("/test", func(ctx *Context) {
		panic(&net.OpError{Err: &os.SyscallError{Err: errors.New("broken pipe")}})
	})

	request, _ := http.NewRequest("GET", "/test", nil)
	s.server.Router.ServeHTTP(s.recorder, request)

	s.Contains(s.recorder.Body.String(), "broken pipe")
}

func TestRecovery(t *testing.T) {
	test.Run(t, new(RecoverySuite))
}
