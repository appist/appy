package pack

import (
	"net/http"
	"strings"
)

func mdwHealthCheck(endpoint string) HandlerFunc {
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
