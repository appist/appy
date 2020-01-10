package appy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type APIOnlyResponseSuite struct {
	TestSuite
	recorder *httptest.ResponseRecorder
}

func (s *APIOnlyResponseSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *APIOnlyResponseSuite) TestAPIOnlyResponse() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Writer.Header().Add("Set-Cookie", "test")
	APIOnlyResponse()(c)
	s.Equal("test", c.Writer.Header().Get("Set-Cookie"))

	c, _ = NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Add("X-API-Only", "true")
	c.Writer.Header().Add("Set-Cookie", "test")
	APIOnlyResponse()(c)
	s.Equal("", c.Writer.Header().Get("Set-Cookie"))
}

func TestAPIOnlyResponseSuite(t *testing.T) {
	RunTestSuite(t, new(APIOnlyResponseSuite))
}
