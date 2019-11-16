package http

import (
	"github.com/appist/appy/internal/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ResponseHeaderFilterSuite struct {
	test.Suite
}

func (s *ResponseHeaderFilterSuite) SetupTest() {
}

func (s *ResponseHeaderFilterSuite) TearDownTest() {
}

func (s *ResponseHeaderFilterSuite) TestResponseHeaderFilter() {
	recorder := httptest.NewRecorder()
	ctx, _ := test.CreateHTTPContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(ctx)
	s.Equal("test", ctx.Writer.Header().Get("Set-Cookie"))

	ctx, _ = test.CreateHTTPContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "true")
	ctx.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(ctx)
	s.Equal("", ctx.Writer.Header().Get("Set-Cookie"))
}

func TestResponseHeaderFilter(t *testing.T) {
	test.RunSuite(t, new(ResponseHeaderFilterSuite))
}
