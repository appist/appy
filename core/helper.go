package core

import (
	"net/http"
)

var apiOnlyHeader = http.CanonicalHeaderKey("x-api-only")

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func IsAPIOnly(ctx *Context) bool {
	if ctx.Request.Header.Get(apiOnlyHeader) == "true" || ctx.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
}
