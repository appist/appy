package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	// Context contains the request information and is meant to be passed through the entire HTTP request.
	Context struct {
		*gin.Context
	}
)

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (c *Context, r *Router) {
	r = newRouter()
	c = &Context{
		&gin.Context{},
	}

	return c, r
}
