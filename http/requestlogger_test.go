package http

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type RequestLoggerSuite struct {
	test.Suite
	assetsMngr *support.AssetsMngr
	config     *support.Config
	logger     *support.Logger
}

func (s *RequestLoggerSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.logger = support.NewLogger()
	s.assetsMngr = support.NewAssetsMngr(nil, "", http.Dir("../support/testdata"))
	s.config = support.NewConfig(s.assetsMngr, s.logger)
}

func (s *RequestLoggerSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *RequestLoggerSuite) TestRequestLogger() {
	recorder := httptest.NewRecorder()
	ctx, _ := NewTestContext(recorder)
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
	test.RunSuite(t, new(RequestLoggerSuite))
}
