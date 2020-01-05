package http

import (
	"github.com/appist/appy/support"
)

var (
	loggerCtxKey = ContextKey("logger")
)

// Logger attaches the logger to the request context.
func Logger(logger *support.Logger) HandlerFunc {
	return func(c *Context) {
		c.Set(loggerCtxKey.String(), logger)
		c.Next()
	}
}
