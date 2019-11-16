package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/internal/test"
)

type HTTPSuite struct {
	test.Suite
}

func (s *HTTPSuite) SetupTest() {
}

func (s *HTTPSuite) TearDownTest() {
}

func (s *HTTPSuite) TestIsAPIOnly() {
	recorder := httptest.NewRecorder()

	ctx, _ := test.CreateHTTPContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "1")
	s.Equal(true, IsAPIOnly(ctx))

	ctx, _ = test.CreateHTTPContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "0")
	s.Equal(false, IsAPIOnly(ctx))
}

func TestHTTPSuite(t *testing.T) {
	test.RunSuite(t, new(HTTPSuite))
}
