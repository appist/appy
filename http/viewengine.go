package http

import (
	"github.com/appist/appy/support"
)

var (
	viewEngineCtxKey = ContextKey("viewEngine")
)

// ViewEngine attaches the jet view engine to the context.
func ViewEngine(viewEngine *support.ViewEngine) HandlerFunc {
	return func(c *Context) {
		c.Set(viewEngineCtxKey.String(), viewEngine)
		c.Next()
	}
}
