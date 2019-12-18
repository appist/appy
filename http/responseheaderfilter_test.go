package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
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
	c, _ := NewTestContext(recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(c)
	s.Equal("test", c.Writer.Header().Get("Set-Cookie"))

	c, _ = NewTestContext(recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Add("X-API-Only", "true")
	c.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(c)
	s.Equal("", c.Writer.Header().Get("Set-Cookie"))
}

func TestResponseHeaderFilter(t *testing.T) {
	test.RunSuite(t, new(ResponseHeaderFilterSuite))
}
