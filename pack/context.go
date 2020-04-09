package pack

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/appist/appy/view"
	"github.com/gin-gonic/gin"
)

var (
	xAPIOnly = http.CanonicalHeaderKey("x-api-only")
)

// Context contains the request information and is meant to be passed through
// the entire HTTP request.
type Context struct {
	*gin.Context
}

// CSRFAuthenticityTemplateField is a template helper for html/template that
// provides an <input> field populated with a CSRF authenticity token.
func (c *Context) CSRFAuthenticityTemplateField() string {
	fieldName := mdwCSRFAuthenticityTemplateFieldName(c)

	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, c.CSRFAuthenticityToken())
}

// CSRFAuthenticityToken returns the CSRF authenticity token for the request.
func (c *Context) CSRFAuthenticityToken() string {
	val, exists := c.Get(mdwCSRFAuthenticityTokenCtxKey.String())
	if exists {
		if token, ok := val.(string); ok {
			return token
		}
	}

	return ""
}

// Deliver sends out the email via SMTP immediately.
func (c *Context) Deliver(mail *mailer.Mail) error {
	ml, _ := c.Get(mdwMailerCtxKey.String())

	if mail.Locale == "" {
		mail.Locale = c.Locale()
	}

	return ml.(*mailer.Engine).Deliver(mail)
}

// HTML renders the HTTP template with the HTTP code and the "text/html" Content-Type header.
func (c *Context) HTML(code int, name string, obj interface{}) {
	viewEngine, _ := c.Get(mdwViewEngineCtxKey.String())
	ve := viewEngine.(*view.Engine)
	ve.AddGlobal("t", func(key string, args ...interface{}) string {
		return c.T(key, args...)
	})

	t, err := ve.GetTemplate(name)
	if err != nil {
		c.Logger().Error(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var w bytes.Buffer
	vars := make(jet.VarMap)
	if err := t.Execute(&w, vars, obj); err != nil {
		c.Logger().Error(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if support.IsReleaseBuild() {
		c.Data(code, "text/html; charset=utf-8", w.Bytes())
		return
	}

	html := strings.ReplaceAll(w.String(), "</body>", liveReloadTpl(c.Request.Host, c.Request.TLS)+"</body>")
	c.Data(code, "text/html; charset=utf-8", []byte(html))
}

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func (c *Context) IsAPIOnly() bool {
	if c.Request.Header.Get(xAPIOnly) == "true" || c.Request.Header.Get(xAPIOnly) == "1" {
		return true
	}

	return false
}

// Locale returns the request context's locale.
func (c *Context) Locale() string {
	locale, exists := c.Get(mdwI18nLocaleCtxKey.String())

	if locale == "" || !exists {
		return "en"
	}

	return locale.(string)
}

// Locales returns the request context's available locales.
func (c *Context) Locales() []string {
	i18n, _ := c.Get(mdwI18nCtxKey.String())

	return i18n.(*support.I18n).Locales()
}

// Logger returns the request context's logger.
func (c *Context) Logger() *support.Logger {
	logger, exists := c.Get(mdwLoggerCtxKey.String())
	if !exists {
		return nil
	}

	return logger.(*support.Logger)
}

// RequestID returns the unique request ID.
func (c *Context) RequestID() string {
	reqID, exists := c.Get(mdwReqIDCtxKey.String())
	if !exists {
		return ""
	}

	return reqID.(string)
}

// SetLocale sets the request's locale.
func (c *Context) SetLocale(locale string) {
	c.Set(mdwI18nLocaleCtxKey.String(), locale)
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

	i18n, _ := c.Get(mdwI18nCtxKey.String())
	return i18n.(*support.I18n).T(key, args...)
}

// ContextKey is the context key with appy namespace.
type ContextKey string

func (c ContextKey) String() string {
	return "appy." + string(c)
}

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}
