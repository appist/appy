package http

import (
	"net/http"
	"strings"
)

// HealthCheck sets up a route to inform the request if the service is healthy.
func HealthCheck(endpoint string) HandlerFunc {
	return func(c *Context) {
		r := c.Request
		if r.Method == "GET" && strings.EqualFold(r.URL.Path, endpoint) {
			c.String(http.StatusOK, "")
			c.Abort()
			return
		}

		c.Next()
	}
}
