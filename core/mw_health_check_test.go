package core

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
)

type HealthCheckSuite struct {
	test.Suite
	config   AppConfig
	logger   *AppLogger
	recorder *httptest.ResponseRecorder
}

func (s *HealthCheckSuite) SetupTest() {
	s.logger, _ = newLogger(newLoggerConfig())
	s.config, _, _ = newConfig(nil, nil, nil, s.logger)
	s.config.HTTPSessionSecrets = [][]byte{[]byte("481e5d98a31585148b8b1dfb6a3c0465")}
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.recorder = httptest.NewRecorder()
}

func (s *HealthCheckSuite) TearDownTest() {
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathMatches() {
	ctx, _ := test.CreateTestContext(s.recorder)
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
	ctx, _ := test.CreateTestContext(s.recorder)
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
	test.Run(t, new(HealthCheckSuite))
}
