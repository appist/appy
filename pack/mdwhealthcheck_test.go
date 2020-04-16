package pack

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwHealthCheckSuite struct {
	test.Suite
	asset    *support.Asset
	config   *support.Config
	logger   *support.Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	server   *Server
	recorder *httptest.ResponseRecorder
}

func (s *mdwHealthCheckSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.recorder = httptest.NewRecorder()
	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger)
}

func (s *mdwHealthCheckSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
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
	mdwHealthCheck("/ping", s.server)(c)

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
	mdwHealthCheck("/health_check", s.server)(c)

	s.NotEqual("text/plain; charset=utf-8", c.Writer.Header().Get("Content-Type"))
}

func TestMdwHealthCheckSuite(t *testing.T) {
	test.Run(t, new(mdwHealthCheckSuite))
}
