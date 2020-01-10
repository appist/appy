package appy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type RequestLoggerSuite struct {
	TestSuite
	logger   *Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
}

func (s *RequestLoggerSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *RequestLoggerSuite) TestRequestLogger() {
	config := &Config{
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
	c.Set(requestIDCtxKey.String(), "1234")

	RequestLogger(config, s.logger)(c)
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
	c.Set(requestIDCtxKey.String(), "1234")

	RequestLogger(config, s.logger)(c)
	s.writer.Flush()
	s.Contains(s.buffer.String(), "[HTTP] 1234 GET 'https://localhost HTTP/2.0' from 127.0.0.1 - 200")
}

func TestRequestLogger(t *testing.T) {
	RunTestSuite(t, new(RequestLoggerSuite))
}
