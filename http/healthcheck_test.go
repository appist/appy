package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
)

type HealthCheckSuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *HealthCheckSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *HealthCheckSuite) TearDownTest() {
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathMatches() {
	ctx, _ := NewTestContext(s.recorder)
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

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathDoesNotMatch() {
	ctx, _ := NewTestContext(s.recorder)
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
	test.RunSuite(t, new(HealthCheckSuite))
}
