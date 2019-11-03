package appy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy"
)

type RequestIDSuite struct {
	appy.TestSuite
}

func (s *RequestIDSuite) SetupTest() {
}

func (s *RequestIDSuite) TearDownTest() {
}

func (s *RequestIDSuite) TestRequestID() {
	recorder := httptest.NewRecorder()
	ctx, _ := appy.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	appy.RequestID()(ctx)
	requestID, _ := ctx.Get("appy.requestID")
	s.NotEqual("", requestID)
}

func TestRequestID(t *testing.T) {
	appy.RunTestSuite(t, new(RequestIDSuite))
}
