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
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{c}, &Router{router}
}

func (c *Context) ginHTML(code int, name string, obj interface{}) {
	c.HTML(code, name, obj)
}
