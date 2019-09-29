package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var apiOnlyHeader = http.CanonicalHeaderKey("x-api-only")

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func IsAPIOnly(c *gin.Context) bool {
	if c.Request.Header.Get(apiOnlyHeader) == "true" || c.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
}
