package http

import (
	"github.com/appist/appy/support"
)

var (
	viewEngineCtxKey = ContextKey("viewEngine")
)

// ViewEngine attaches the jet view engine to the context.
func ViewEngine(assets *support.Assets, viewFuncs map[string]interface{}) HandlerFunc {
	return func(c *Context) {
		ve := support.NewViewEngine(assets)

		if viewFuncs != nil {
			for name, vf := range viewFuncs {
				ve.AddGlobal(name, vf)
			}
		}

		c.Set(viewEngineCtxKey.String(), ve)
		c.Next()
	}
}
