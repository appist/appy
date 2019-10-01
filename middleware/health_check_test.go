package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type HealthCheckSuiteT struct {
	test.SuiteT
	Config   *support.ConfigT
	Recorder *httptest.ResponseRecorder
}

func (s *HealthCheckSuiteT) SetupTest() {
	s.Config = support.Config
	s.Recorder = httptest.NewRecorder()
}

func (s *HealthCheckSuiteT) TestCorrectResponseIfRequestPathMatches() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	HealthCheck("/ping")(ctx)

	s.Equal("text/plain; charset=utf-8", ctx.Writer.Header().Get("Content-Type"))
	s.Equal(200, ctx.Writer.Status())
}

func (s *HealthCheckSuiteT) TestCorrectResponseIfRequestPathDoesNotMatch() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	HealthCheck("/health_check")(ctx)

	s.NotEqual("text/plain; charset=utf-8", ctx.Writer.Header().Get("Content-Type"))
}

func TestHealthCheck(t *testing.T) {
	test.Run(t, new(HealthCheckSuiteT))
}
