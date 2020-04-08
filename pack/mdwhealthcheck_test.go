package pack

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
)

type mdwHealthCheckSuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *mdwHealthCheckSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *mdwHealthCheckSuite) TestCorrectResponseIfRequestPathMatches() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	mdwHealthCheck("/ping")(c)

	s.Equal("text/plain; charset=utf-8", c.Writer.Header().Get("Content-Type"))
	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *mdwHealthCheckSuite) TestCorrectResponseIfRequestPathDoesNotMatch() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		URL: &url.URL{
			Path: "/ping",
		},
	}
	mdwHealthCheck("/health_check")(c)

	s.NotEqual("text/plain; charset=utf-8", c.Writer.Header().Get("Content-Type"))
}

func TestMdwHealthCheckSuite(t *testing.T) {
	test.Run(t, new(mdwHealthCheckSuite))
}
