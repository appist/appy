package appy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type RequestIDSuite struct {
	TestSuite
	recorder *httptest.ResponseRecorder
}

func (s *RequestIDSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *RequestIDSuite) TestRequestID() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	s.Empty(c.RequestID())
	RequestID()(c)
	s.NotEmpty(c.RequestID())
}

func TestRequestID(t *testing.T) {
	RunTestSuite(t, new(RequestIDSuite))
}
