package pack

import (
	"net/http"
	"strings"
)

func mdwHealthCheck(endpoint string, server *Server) HandlerFunc {
	server.mdwRoutes = append(server.mdwRoutes, Route{
		Method:      "GET",
		Path:        endpoint,
		Handler:     "github.com/appist/appy/pack.mdwHealthCheck",
		HandlerFunc: nil,
	})

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
