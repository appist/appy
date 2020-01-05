package http

import (
	"github.com/appist/appy/support"
)

var (
	viewEngineCtxKey = ContextKey("viewEngine")
)

// ViewEngine attaches the jet view engine to the context.
func ViewEngine(assets *support.Assets, config *support.Config, logger *support.Logger, viewFuncs map[string]interface{}) HandlerFunc {
	return func(c *Context) {
		ve := support.NewViewEngine(assets, config, logger)
		ve.SetGlobalFuncs(viewFuncs)

		c.Set(viewEngineCtxKey.String(), ve)
		c.Next()
	}
}
