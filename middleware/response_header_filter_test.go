package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
	"github.com/gin-gonic/gin"
)

func TestResponseHeaderFilter(t *testing.T) {
	assert := test.NewAssert(t)
	recorder := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(ctx)
	assert.Equal("test", ctx.Writer.Header().Get("Set-Cookie"))

	ctx, _ = gin.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "true")
	ctx.Writer.Header().Add("Set-Cookie", "test")
	ResponseHeaderFilter()(ctx)
	assert.Equal("", ctx.Writer.Header().Get("Set-Cookie"))
}
