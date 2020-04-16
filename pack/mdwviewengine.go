package pack

import (
	"github.com/appist/appy/support"
	"github.com/appist/appy/view"
)

var (
	mdwViewEngineCtxKey = ContextKey("mdwViewEngine")
)

func mdwViewEngine(asset *support.Asset, config *support.Config, logger *support.Logger, viewFuncs map[string]interface{}) HandlerFunc {
	return func(c *Context) {
		ve := view.NewEngine(asset, config, logger)
		ve.SetGlobalFuncs(viewFuncs)

		c.Set(mdwViewEngineCtxKey.String(), ve)
		c.Next()
	}
}
