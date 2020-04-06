package pack

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context contains the request information and is meant to be passed through
// the entire HTTP request.
type Context struct {
	*gin.Context
}

// ContextKey is the context key with appy namespace.
type ContextKey string

func (c ContextKey) String() string {
	return "appy." + string(c)
}

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}
