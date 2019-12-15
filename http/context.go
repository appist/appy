package http

import "github.com/gin-gonic/gin"

type (
	// Context contains the request information and is meant to be passed through the entire HTTP request.
	Context struct {
		*gin.Context
	}
)

var (
	// CreateHTTPContext returns a fresh router w/ context for testing purposes.
	CreateHTTPContext = gin.CreateTestContext
)
