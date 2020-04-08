package pack

import (
	"errors"
	"strings"

	"github.com/appist/appy/support"
	"github.com/gorilla/securecookie"
)

var (
	mdwCSRFSecureCookie    *securecookie.SecureCookie
	mdwCSRFFieldNameCtxKey = ContextKey("csrfFieldName")
	mdwCSRFSkipCheckCtxKey = ContextKey("csrfSkipCheck")
	mdwCSRFTokenCtxKey     = ContextKey("csrfToken")
	mdwCSRFSafeMethods     = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	errCSRFNoReferer       = errors.New("the request referer is missing")
	errCSRFBadReferer      = errors.New("the request referer is invalid")
	errCSRFNoToken         = errors.New("the CSRF token is missing")
	errCSRFBadToken        = errors.New("the CSRF token is invalid")
)

func mdwCSRF(config *support.Config, logger *support.Logger) HandlerFunc {
	mdwCSRFSecureCookie = securecookie.New(config.HTTPCSRFSecret, nil)
	mdwCSRFSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	mdwCSRFSecureCookie.MaxAge(config.HTTPCSRFCookieMaxAge)

	return func(c *Context) {
		mdwCSRFHandler(c, config, logger)
	}
}

// CSRFSkipCheck skips the CSRF check for the request.
func CSRFSkipCheck() HandlerFunc {
	return func(c *Context) {
		c.Set(mdwCSRFSkipCheckCtxKey.String(), true)
		c.Next()
	}
}

func mdwCSRFHandler(c *Context, config *support.Config, logger *support.Logger) {
	if c.IsAPIOnly() {
		c.Set(mdwCSRFSkipCheckCtxKey.String(), true)
	}

	skipCheck, exists := c.Get(mdwCSRFSkipCheckCtxKey.String())
	if exists && skipCheck.(bool) {
		c.Next()
		return
	}
}

func mdwCSRFTemplateFieldName(c *Context) string {
	fieldName, exists := c.Get(mdwCSRFFieldNameCtxKey.String())

	if fieldName == "" || !exists {
		fieldName = "authenticity_token"
	}

	return strings.ToLower(fieldName.(string))
}
