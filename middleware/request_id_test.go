package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

func TestRequestID(t *testing.T) {
	assert := test.NewAssert(t)
	recorder := httptest.NewRecorder()
	ctx, _ := test.CreateContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	RequestID()(ctx)
	requestID, _ := ctx.Get(requestIDCtxKey)
	assert.NotEqual("", requestID)
}
