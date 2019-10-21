package core

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
)

type RequestLoggerSuite struct {
	test.Suite
	config AppConfig
	logger *AppLogger
}

func (s *RequestLoggerSuite) SetupTest() {
	s.logger, _ = newLogger(newLoggerConfig())
	s.config, _, _ = newConfig(nil, nil, s.logger)
	s.config.HTTPSessionSecrets = [][]byte{[]byte("481e5d98a31585148b8b1dfb6a3c0465")}
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
}

func (s *RequestLoggerSuite) TearDownTest() {
}

func (s *RequestLoggerSuite) TestRequestLogger() {
	recorder := httptest.NewRecorder()
	ctx, _ := test.CreateTestContext(recorder)
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
	test.Run(t, new(RequestLoggerSuite))
}
