package appy_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/appist/appy"
)

type HealthCheckSuite struct {
	appy.TestSuite
	config   *appy.Config
	logger   *appy.Logger
	recorder *httptest.ResponseRecorder
}

func (s *HealthCheckSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.logger = appy.NewLogger(appy.DebugBuild)
	s.config = appy.NewConfig(appy.DebugBuild, s.logger, nil)
	s.recorder = httptest.NewRecorder()
}

func (s *HealthCheckSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathMatches() {
	ctx, _ := appy.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	appy.HealthCheck("/ping")(ctx)

	s.Equal("text/plain; charset=utf-8", ctx.Writer.Header().Get("Content-Type"))
	s.Equal(200, ctx.Writer.Status())
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathDoesNotMatch() {
	ctx, _ := appy.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	appy.HealthCheck("/health_check")(ctx)

	s.NotEqual("text/plain; charset=utf-8", ctx.Writer.Header().Get("Content-Type"))
}

func TestHealthCheck(t *testing.T) {
	appy.RunTestSuite(t, new(HealthCheckSuite))
}
