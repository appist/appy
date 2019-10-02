package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

func TestIsAPIOnly(t *testing.T) {
	assert := test.NewAssert(t)
	recorder := httptest.NewRecorder()

	ctx, _ := test.CreateContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "1")
	assert.Equal(true, IsAPIOnly(ctx))

	ctx, _ = test.CreateContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "0")
	assert.Equal(false, IsAPIOnly(ctx))
}
