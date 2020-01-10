package appy

var (
	loggerCtxKey = ContextKey("logger")
)

// AttachLogger attaches the logger to the request context.
func AttachLogger(logger *Logger) HandlerFunc {
	return func(c *Context) {
		c.Set(loggerCtxKey.String(), logger)
		c.Next()
	}
}
