package core

import (
	"net/http"
	"strings"
)

// HealthCheck is middleware to setup a path like `/health_check` that load balancers or uptime testing external
// services can make a request before hitting any routes.
func HealthCheck(endpoint string) HandlerFunc {
	return func(ctx *Context) {
		r := ctx.Request
		if r.Method == "GET" && strings.EqualFold(r.URL.Path, endpoint) {
			ctx.String(http.StatusOK, "")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
