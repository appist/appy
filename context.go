package appy

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet"
	"github.com/gin-gonic/gin"
)

const (
	// LiveReloadWSPort is the server-side live reload port over HTTP.
	LiveReloadWSPort = "12450"

	// LiveReloadWSSPort is the server-side live reload port over HTTPS.
	LiveReloadWSSPort = "12451"

	// LiveReloadPath is the server-side live reload path.
	LiveReloadPath = "/reload"
)

var (
	apiOnlyHeader = http.CanonicalHeaderKey("x-api-only")
)

// Context contains the request information and is meant to be passed through the entire HTTP request.
type Context struct {
	*gin.Context
}

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}

// CSRFTemplateField is a template helper for html/template that provides an <input> field populated with a CSRF token.
func (c *Context) CSRFTemplateField() string {
	fieldName := csrfTemplateFieldName(c)

	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, c.CSRFToken())
}

// CSRFToken returns the CSRF token for the request.
func (c *Context) CSRFToken() string {
	val, exists := c.Get(csrfTokenCtxKey.String())
	if exists {
		if token, ok := val.(string); ok {
			return token
		}
	}

	return ""
}

// DeliverMail sends out the email via SMTP immediately.
func (c *Context) DeliverMail(mail Mail) error {
	mailer, _ := c.Get(mailerCtxKey.String())

	if mail.Locale == "" {
		mail.Locale = c.Locale()
	}

	return mailer.(*Mailer).Deliver(mail)
}

// HTML renders the HTTP template with the HTTP code and the "text/html" Content-Type header.
func (c *Context) HTML(code int, name string, obj interface{}) {
	ve, _ := c.Get(viewEngineCtxKey.String())
	viewEngine := ve.(*ViewEngine)
	viewEngine.AddGlobal("t", func(key string, args ...interface{}) string {
		return c.T(key, args...)
	})

	t, err := viewEngine.GetTemplate(name)
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

	if IsReleaseBuild() {
		c.Data(code, "text/html; charset=utf-8", w.Bytes())
		return
	}

	html := strings.ReplaceAll(w.String(), "</body>", c.LiveReloadTpl()+"</body>")
	c.Data(code, "text/html; charset=utf-8", []byte(html))
}

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func (c *Context) IsAPIOnly() bool {
	if c.Request.Header.Get(apiOnlyHeader) == "true" || c.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
}

// LiveReloadTpl returns the live reload template that auto refresh the browser after the server is re-compiled.
func (c *Context) LiveReloadTpl() string {
	protocol := "ws"
	port := LiveReloadWSPort

	if c.Request.TLS != nil {
		protocol = "wss"
		port = LiveReloadWSSPort
	}

	splits := strings.Split(c.Request.Host, ":")
	url := protocol + `://` + splits[0] + ":" + port + LiveReloadPath

	return `<script>function b(a){var c=new WebSocket(a);c.onclose=function(){setTimeout(function(){b(a)},2E3)};` +
		`c.onmessage=function(){location.reload()}}try{if(window.WebSocket)try{b("` + url + `")}catch(a){console.error(a)}` +
		`else console.log("Your browser does not support WebSocket.")}catch(a){console.error(a)};</script>`
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

	return i18n.(*I18n).Locales()
}

// Logger returns the request context's logger.
func (c *Context) Logger() *Logger {
	logger, exists := c.Get(loggerCtxKey.String())
	if !exists {
		return nil
	}

	return logger.(*Logger)
}

// RequestID returns the unique request ID.
func (c *Context) RequestID() string {
	reqID, exists := c.Get(requestIDCtxKey.String())
	if !exists {
		return ""
	}

	return reqID.(string)
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
	return i18n.(*I18n).T(key, args...)
}

// DefaultHTML uses the gin's default HTML method which doesn't use Jet template engine and is only meant for internal
// use.
func (c *Context) defaultHTML(code int, name string, obj interface{}) {
	c.Context.HTML(code, name, obj)
}

// ContextKey is the context key with appy namespace.
type ContextKey string

func (c ContextKey) String() string {
	return "appy." + string(c)
}
