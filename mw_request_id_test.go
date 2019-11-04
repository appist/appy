package appy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type RequestIDSuite struct {
	TestSuite
}

func (s *RequestIDSuite) SetupTest() {
}

func (s *RequestIDSuite) TearDownTest() {
}

func (s *RequestIDSuite) TestRequestID() {
	recorder := httptest.NewRecorder()
	ctx, _ := CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	RequestID()(ctx)
	requestID, _ := ctx.Get("requestID")
	s.NotEqual("", requestID)
}

func TestRequestID(t *testing.T) {
	RunTestSuite(t, new(RequestIDSuite))
}
