package appy

import (
	"fmt"
	"net/http"

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

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func (c *Context) IsAPIOnly() bool {
	if c.Request.Header.Get(apiOnlyHeader) == "true" || c.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
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
