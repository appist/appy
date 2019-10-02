package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/gin-gonic/gin"
)

func TestFilterParams(t *testing.T) {
	assert := test.NewAssert(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Host:       "0.0.0.0",
		Method:     "GET",
		Proto:      "HTTP/2.0",
		RemoteAddr: "0.0.0.0",
		TLS:        &tls.ConnectionState{},
		URL:        &url.URL{},
	}
	RequestLogger(support.Config)(ctx)

	host := "http://0.0.0.0"
	query := "username=user&password=secret"
	request := &http.Request{
		RequestURI: host + "?" + query,
	}
	request.URL = &url.URL{
		RawQuery: query,
	}

	assert.Contains(filterParams(request, support.Config), "password=[FILTERED]")
}
