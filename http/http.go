package http

import "net/http"

type (
	// ContextKey is the context key with appy namespace.
	ContextKey string
)

var (
	apiOnlyHeader = http.CanonicalHeaderKey("x-api-only")
)

func (c ContextKey) String() string {
	return "appy." + string(c)
}

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func IsAPIOnly(c *Context) bool {
	if c.Request.Header.Get(apiOnlyHeader) == "true" || c.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
}
