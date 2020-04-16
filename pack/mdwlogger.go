package pack

import "github.com/appist/appy/support"

var (
	mdwLoggerCtxKey = ContextKey("mdwLogger")
)

func mdwLogger(logger *support.Logger) HandlerFunc {
	return func(c *Context) {
		c.Set(mdwLoggerCtxKey.String(), logger)
		c.Next()
	}
}
