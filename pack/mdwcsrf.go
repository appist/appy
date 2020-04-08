package pack

import (
	"errors"
	"strings"

	"github.com/appist/appy/support"
	"github.com/gorilla/securecookie"
)

var (
	csrfSecureCookie    *securecookie.SecureCookie
	csrfFieldNameCtxKey = ContextKey("csrfFieldName")
	csrfSkipCheckCtxKey = ContextKey("csrfSkipCheck")
	csrfTokenCtxKey     = ContextKey("csrfToken")
	csrfSafeMethods     = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	errCsrfNoReferer    = errors.New("the request referer is missing")
	errCsrfBadReferer   = errors.New("the request referer is invalid")
	errCsrfNoToken      = errors.New("the CSRF token is missing")
	errCsrfBadToken     = errors.New("the CSRF token is invalid")
)

func mdwCSRF(config *support.Config, logger *support.Logger) HandlerFunc {
	csrfSecureCookie = securecookie.New(config.HTTPCSRFSecret, nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(config.HTTPCSRFCookieMaxAge)

	return func(c *Context) {
		csrfHandler(c, config, logger)
	}
}

// CSRFSkipCheck skips the CSRF check for the request.
func CSRFSkipCheck() HandlerFunc {
	return func(c *Context) {
		c.Set(csrfSkipCheckCtxKey.String(), true)
		c.Next()
	}
}

func csrfHandler(c *Context, config *support.Config, logger *support.Logger) {
	if c.IsAPIOnly() {
		c.Set(csrfSkipCheckCtxKey.String(), true)
	}

	skipCheck, exists := c.Get(csrfSkipCheckCtxKey.String())
	if exists && skipCheck.(bool) {
		c.Next()
		return
	}
}

func csrfTemplateFieldName(c *Context) string {
	fieldName, exists := c.Get(csrfFieldNameCtxKey.String())

	if fieldName == "" || !exists {
		fieldName = "authenticity_token"
	}

	return strings.ToLower(fieldName.(string))
}
