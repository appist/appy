package appy

var (
	viewEngineCtxKey = ContextKey("viewEngine")
)

// AttachViewEngine attaches the jet view engine to the context.
func AttachViewEngine(asset *Asset, config *Config, logger *Logger, viewFuncs map[string]interface{}) HandlerFunc {
	return func(c *Context) {
		ve := NewViewEngine(asset, config, logger)
		ve.SetGlobalFuncs(viewFuncs)

		c.Set(viewEngineCtxKey.String(), ve)
		c.Next()
	}
}
