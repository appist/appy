package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/html"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

type RecoverySuiteT struct {
	test.SuiteT
	Config   *support.ConfigT
	Recorder *httptest.ResponseRecorder
	Renderer multitemplate.Renderer
	Router   *gin.Engine
}

func (s *RecoverySuiteT) SetupTest() {
	s.Config = &support.ConfigT{}
	support.Copy(&s.Config, &support.Config)
	s.Recorder = httptest.NewRecorder()
	s.Renderer = multitemplate.NewRenderer()
	s.Renderer.AddFromString("error/500", html.ErrorTpl500())
	s.Router = gin.New()
	s.Router.HTMLRender = s.Renderer
	s.Router.Use(Recovery())
}

func (s *RecoverySuiteT) TestPanicRenders500() {
	s.Router.Use(SessionManager(s.Config))
	s.Router.GET("/test", func(ctx *gin.Context) {
		s := DefaultSession(ctx)
		s.Set("username", "dummy")
		panic(errors.New("error"))
	})

	request, _ := http.NewRequest("GET", "/test?age=10", nil)
	request.Header.Set("X-Testing", "1")
	s.Router.ServeHTTP(s.Recorder, request)

	s.Equal(http.StatusInternalServerError, s.Recorder.Code)
	s.Contains(s.Recorder.Body.String(), "<title>500 Internal Server Error</title>")
	s.Contains(s.Recorder.Body.String(), "username: dummy")
	s.Contains(s.Recorder.Body.String(), "X-Testing: 1")
	s.Contains(s.Recorder.Body.String(), "age: 10")
}

func (s *RecoverySuiteT) TestBrokenPipeErrorHandling() {
	s.Router.GET("/test", func(ctx *gin.Context) {
		panic(&net.OpError{Err: &os.SyscallError{Err: errors.New("broken pipe")}})
	})

	request, _ := http.NewRequest("GET", "/test", nil)
	s.Router.ServeHTTP(s.Recorder, request)

	s.Contains(s.Recorder.Body.String(), "broken pipe")
}

func TestRecovery(t *testing.T) {
	test.Run(t, new(RecoverySuiteT))
}
