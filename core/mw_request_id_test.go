package core

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
	ctx, _ := test.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	RequestID()(ctx)
	requestID, _ := ctx.Get(requestIDCtxKey)
	s.NotEqual("", requestID)
}

func TestRequestID(t *testing.T) {
	test.Run(t, new(RequestIDSuite))
}
