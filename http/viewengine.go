package http

import (
	"github.com/appist/appy/support"
)

var (
	viewEngineCtxKey = ContextKey("viewEngine")
)

// ViewEngine attaches the jet view engine to the context.
func ViewEngine(assets *support.Assets) HandlerFunc {
	return func(c *Context) {
		c.Set(viewEngineCtxKey.String(), support.NewViewEngine(assets))
		c.Next()
	}
}
