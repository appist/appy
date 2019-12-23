package http

import (
	"bytes"
	"net/http"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
)

type (
	// Context contains the request information and is meant to be passed through the entire HTTP request.
	Context struct {
		*gin.Context
	}
)

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}

// HTML renders the HTTP template with the HTTP code and the "text/html" Content-Type header.
func (c *Context) HTML(code int, name string, obj interface{}) {
	viewEngine, _ := c.Get(viewEngineCtxKey.String())
	t, err := viewEngine.(*support.ViewEngine).GetTemplate(name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var w bytes.Buffer
	vars := make(jet.VarMap)
	if err = t.Execute(&w, vars, obj); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Data(code, "text/html", w.Bytes())
}

// Locale returns the request context's locale.
func (c *Context) Locale() string {
	locale, exists := c.Get(i18nLocaleCtxKey.String())

	if locale == "" || !exists {
		return "en"
	}

	return locale.(string)
}

// Locales returns the request context's available locales.
func (c *Context) Locales() []string {
	i18n, _ := c.Get(i18nCtxKey.String())

	return i18n.(*support.I18n).Locales()
}

// Session returns the session in the request context.
func (c *Context) Session() Sessioner {
	s, exists := c.Get(sessionManagerCtxKey.String())
	if !exists {
		return nil
	}

	return s.(Sessioner)
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

	i18n, _ := c.Get(i18nCtxKey.String())
	return i18n.(*support.I18n).T(key, args...)
}

func (c *Context) ginHTML(code int, name string, obj interface{}) {
	c.Context.HTML(code, name, obj)
}
