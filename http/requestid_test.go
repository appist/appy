package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type RequestIDSuite struct {
	test.Suite
}

func (s *RequestIDSuite) SetupTest() {
}

func (s *RequestIDSuite) TearDownTest() {
}

func (s *RequestIDSuite) TestRequestID() {
	recorder := httptest.NewRecorder()
	c, _ := NewTestContext(recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	RequestID()(c)
	requestID, _ := c.Get("requestID")
	s.NotEqual("", requestID)
}

func TestRequestID(t *testing.T) {
	test.RunSuite(t, new(RequestIDSuite))
}
