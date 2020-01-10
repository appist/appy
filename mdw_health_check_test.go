package appy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type HealthCheckSuite struct {
	TestSuite
	recorder *httptest.ResponseRecorder
}

func (s *HealthCheckSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathMatches() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	HealthCheck("/ping")(c)

	s.Equal("text/plain; charset=utf-8", c.Writer.Header().Get("Content-Type"))
	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *HealthCheckSuite) TestCorrectResponseIfRequestPathDoesNotMatch() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	HealthCheck("/health_check")(c)

	s.NotEqual("text/plain; charset=utf-8", c.Writer.Header().Get("Content-Type"))
}

func TestHealthCheckSuite(t *testing.T) {
	RunTestSuite(t, new(HealthCheckSuite))
}
