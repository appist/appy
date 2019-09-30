package middleware

import (
	"github.com/gin-gonic/gin"
)

// ResponseHeaderFilter is a middleware that removes `X-CSRF-Token` and `Set-Cookie` response header when the request header
// `X-API-Only: 1` is received.
func ResponseHeaderFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAPIOnly(c) == true {
			c.Writer.Header().Del("Set-Cookie")
			c.Writer.Header().Del("X-CSRF-Token")
		}

		c.Next()
	}
}
