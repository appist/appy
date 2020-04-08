package pack

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwReqLoggerSuite struct {
	test.Suite
	logger   *support.Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
}

func (s *mdwReqLoggerSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *mdwReqLoggerSuite) TestRequestLogger() {
	config := &support.Config{
		HTTPLogFilterParameters: []string{"password"},
	}
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Method:     "GET",
		Proto:      "HTTP/2.0",
		Host:       "localhost",
		RemoteAddr: "127.0.0.1",
		TLS:        &tls.ConnectionState{},
		URL: &url.URL{
			RawQuery: "username=user&password=secret",
		},
	}
	c.Set(mdwReqIDCtxKey.String(), "1234")

	mdwReqLogger(config, s.logger)(c)
	s.writer.Flush()
	s.Contains(s.buffer.String(), "[HTTP] 1234 GET 'https://localhost")
	s.Contains(s.buffer.String(), "username=user")
	s.Contains(s.buffer.String(), "password=[FILTERED]")
	s.Contains(s.buffer.String(), " HTTP/2.0' from 127.0.0.1 - 200")

	c, _ = NewTestContext(s.recorder)
	c.Request = &http.Request{
		Method:     "GET",
		Proto:      "HTTP/2.0",
		Host:       "localhost",
		RemoteAddr: "127.0.0.1",
		TLS:        &tls.ConnectionState{},
		URL:        &url.URL{},
	}
	c.Set(mdwReqIDCtxKey.String(), "1234")

	mdwReqLogger(config, s.logger)(c)
	s.writer.Flush()
	s.Contains(s.buffer.String(), "[HTTP] 1234 GET 'https://localhost HTTP/2.0' from 127.0.0.1 - 200")
}

func TestMdwReqLoggerSuite(t *testing.T) {
	test.Run(t, new(mdwReqLoggerSuite))
}
