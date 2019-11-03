package appy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type RequestLoggerSuite struct {
	TestSuite
	config *Config
	logger *Logger
}

func (s *RequestLoggerSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.logger = NewLogger(DebugBuild)
	s.config = NewConfig(DebugBuild, s.logger, nil)
}

func (s *RequestLoggerSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *RequestLoggerSuite) TestRequestLogger() {
	recorder := httptest.NewRecorder()
	ctx, _ := CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Host:       "localhost",
		Method:     "GET",
		Proto:      "HTTP/2.0",
		RemoteAddr: "localhost",
		TLS:        &tls.ConnectionState{},
		URL:        &url.URL{},
	}
	RequestLogger(s.config, s.logger)(ctx)

	host := "http://localhost"
	query := "username=user&password=secret"
	request := &http.Request{
		RequestURI: host + "?" + query,
	}
	request.URL = &url.URL{
		RawQuery: query,
	}

	s.Contains(filterParams(request, s.config), "password=[FILTERED]")
}

func TestRequestLogger(t *testing.T) {
	RunTestSuite(t, new(RequestLoggerSuite))
}
