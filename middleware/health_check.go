package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HealthCheck is middleware to setup a path like `/health_check` that load balancers or uptime testing external
// services can make a request before hitting any routes.
func HealthCheck(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		if r.Method == "GET" && strings.EqualFold(r.URL.Path, endpoint) {
			c.String(http.StatusOK, "")
			c.Abort()
			return
		}

		c.Next()
	}
}
