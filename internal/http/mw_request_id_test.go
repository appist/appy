package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/internal/test"
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
	ctx, _ := test.CreateHTTPContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	RequestID()(ctx)
	requestID, _ := ctx.Get("requestID")
	s.NotEqual("", requestID)
}

func TestRequestID(t *testing.T) {
	test.RunSuite(t, new(RequestIDSuite))
}
