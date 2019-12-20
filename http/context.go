package http

import (
	"net/http"

	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
)

type (
	// Context contains the request information and is meant to be passed through the entire HTTP request.
	Context struct {
		*gin.Context
		i18n *support.I18n
	}
)

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}

// Locale returns the request context's locale.
func (c *Context) Locale() string {
	locale, exists := c.Get(i18nLocaleCtxKey.String())

	if locale == "" || !exists {
		return "en"
	}

	return locale.(string)
}

// SetLocale sets the request context's locale.
func (c *Context) SetLocale(locale string) {
	c.Set(i18nLocaleCtxKey.String(), locale)
}

// T translates a message based on the given key.
func (c *Context) T(key string, args ...interface{}) string {
	var locale string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			locale = v
		}
	}

	if locale == "" {
		args = append(args, c.Locale())
	}

	return c.i18n.T(key, args...)
}

func (c *Context) ginHTML(code int, name string, obj interface{}) {
	c.HTML(code, name, obj)
}
