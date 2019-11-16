package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	// H is a type alias to gin.H.
	H = gin.H

	// HandlerFunc is a type alias to gin.HandlerFunc.
	HandlerFunc = gin.HandlerFunc

	// Context contains the HTTP request information.
	Context = gin.Context

	// Middleware is the middleware list attached to the server.
	Middleware = gin.HandlersChain

	// Router manages the routing logic.
	Router = gin.Engine

	// RouterGroup can be used to group routes.
	RouterGroup = gin.RouterGroup

	// RouteInfo provides the information about a route.
	RouteInfo = gin.RouteInfo

	// Routes is a type alias to gin.IRoutes.
	Routes = gin.IRoutes

	// ContextKey is the context key with appy namespace.
	ContextKey string
)

var (
	apiOnlyHeader = http.CanonicalHeaderKey("x-api-only")
)

func (c ContextKey) String() string {
	return "appy." + string(c)
}

// IsAPIOnly checks if a request is API only based on `X-API-Only` request header.
func IsAPIOnly(ctx *Context) bool {
	if ctx.Request.Header.Get(apiOnlyHeader) == "true" || ctx.Request.Header.Get(apiOnlyHeader) == "1" {
		return true
	}

	return false
}
