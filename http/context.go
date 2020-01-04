package http

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
)

type (
	// Context contains the request information and is meant to be passed through the entire HTTP request.
	Context struct {
		*gin.Context
	}
)

const (
	LiveReloadWSPort  = "12450"
	LiveReloadWSSPort = "12451"
	LiveReloadPath    = "/reload"
)

// NewTestContext returns a fresh router w/ context for testing purposes.
func NewTestContext(w http.ResponseWriter) (*Context, *Router) {
	c, router := gin.CreateTestContext(w)

	return &Context{Context: c}, &Router{router}
}

// HTML renders the HTTP template with the HTTP code and the "text/html" Content-Type header.
func (c *Context) HTML(code int, name string, obj interface{}) {
	ve, _ := c.Get(viewEngineCtxKey.String())
	viewEngine := ve.(*support.ViewEngine)
	viewEngine.AddGlobal("t", func(key string, args ...interface{}) string {
		return c.T(key, args...)
	})

	t, err := viewEngine.GetTemplate(name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var w bytes.Buffer
	vars := make(jet.VarMap)
	if err := t.Execute(&w, vars, obj); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if support.IsReleaseBuild() {
		c.Data(code, "text/html; charset=utf-8", w.Bytes())
		return
	}

	html := strings.ReplaceAll(w.String(), "</body>", c.LiveReloadTpl()+"</body>")
	c.Data(code, "text/html; charset=utf-8", []byte(html))
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

// SendEmail sends out the email via SMTP immediately.
func (c *Context) SendEmail(email mailer.Email) error {
	m, _ := c.Get(mailerCtxKey.String())

	if email.Locale == "" {
		email.Locale = c.Locale()
	}

	return m.(*mailer.Mailer).Send(email)
}

// SendEmailWithTLS sends out the email via secure SMTP immediately.
func (c *Context) SendEmailWithTLS(email mailer.Email) error {
	m, _ := c.Get(mailerCtxKey.String())

	if email.Locale == "" {
		email.Locale = c.Locale()
	}

	return m.(*mailer.Mailer).SendWithTLS(email)
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

// DefaultHTML uses the gin's default HTML method which doesn't use Jet template engine and is only meant for internal
// use.
func (c *Context) DefaultHTML(code int, name string, obj interface{}) {
	c.Context.HTML(code, name, obj)
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
