package appy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy"
)

type ResponseHeaderFilterSuite struct {
	appy.TestSuite
}

func (s *ResponseHeaderFilterSuite) SetupTest() {
}

func (s *ResponseHeaderFilterSuite) TearDownTest() {
}

func (s *ResponseHeaderFilterSuite) TestResponseHeaderFilter() {
	recorder := httptest.NewRecorder()
	ctx, _ := appy.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Writer.Header().Add("Set-Cookie", "test")
	appy.ResponseHeaderFilter()(ctx)
	s.Equal("test", ctx.Writer.Header().Get("Set-Cookie"))

	ctx, _ = appy.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "true")
	ctx.Writer.Header().Add("Set-Cookie", "test")
	appy.ResponseHeaderFilter()(ctx)
	s.Equal("", ctx.Writer.Header().Get("Set-Cookie"))
}

func TestResponseHeaderFilter(t *testing.T) {
	appy.RunTestSuite(t, new(ResponseHeaderFilterSuite))
}
