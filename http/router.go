package http

import (
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

type (
	// H is a shortcut for map[string]interface{}.
	H map[string]interface{}

	// HandlerFunc defines the handler used by middleware as return value.
	HandlerFunc func(*Context)

	// HandlersChain defines the HandlerFunc array.
	HandlersChain []HandlerFunc

	// Router maintains the routing logic.
	Router struct {
		*gin.Engine
	}

	// Route represents a request route's specification which contains method and path and its handler.
	Route struct {
		Method      string
		Path        string
		Handler     string
		HandlerFunc HandlerFunc
	}

	// RouteGroup is associated with a prefix and an array of handlers.
	RouteGroup struct {
		*gin.RouterGroup
	}

	// Routes defines the array of Route.
	Routes []Route

	// Routerer defines all router handler interface.
	Routerer interface {
		Use(...HandlerFunc) Routerer

		Handle(string, string, ...HandlerFunc) Routerer
		Any(string, ...HandlerFunc) Routerer
		GET(string, ...HandlerFunc) Routerer
		POST(string, ...HandlerFunc) Routerer
		DELETE(string, ...HandlerFunc) Routerer
		PATCH(string, ...HandlerFunc) Routerer
		PUT(string, ...HandlerFunc) Routerer
		OPTIONS(string, ...HandlerFunc) Routerer
		HEAD(string, ...HandlerFunc) Routerer

		StaticFile(string, string) Routerer
		Static(string, string) Routerer
		StaticFS(string, http.FileSystem) Routerer
	}
)

func newRouter() *Router {
	r := &Router{
		gin.New(),
	}
	r.AppEngine = true
	r.ForwardedByClientIP = true
	r.HandleMethodNotAllowed = true
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true
	r.UnescapePathValues = true
	r.UseRawPath = false

	// Initialize the error templates.
	renderer := multitemplate.NewRenderer()
	renderer.AddFromString("error/404", errorTpl404())
	renderer.AddFromString("error/500", errorTpl500())
	renderer.AddFromString("default/welcome", welcomeTpl())
	r.HTMLRender = renderer

	return r
}

func (r *Router) routes() []Route {
	routes := []Route{}

	for _, route := range r.Routes() {
		routes = append(routes, Route{
			Method:      route.Method,
			Path:        route.Path,
			Handler:     route.Handler,
			HandlerFunc: func(c *Context) { route.HandlerFunc(c.Context) },
		})
	}

	return routes
}

// Group creates a new route group. You should add all the routes that have common middlewares or the same path prefix.
func (r *Router) Group(path string, handlers ...HandlerFunc) *RouteGroup {
	group := r.RouterGroup.Group(path, wrapHandlers(handlers...)...)

	return &RouteGroup{
		group,
	}
}

// Handle registers a new request handle with the method, given path and middleware.
func (r *Router) Handle(method, path string, handlers ...HandlerFunc) {
	r.Engine.Handle(method, path, wrapHandlers(handlers...)...)
}

// Any registers a route that matches all the HTTP methods, i.e. GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE,
// CONNECT, TRACE.
func (r *Router) Any(path string, handlers ...HandlerFunc) {
	r.Engine.Any(path, wrapHandlers(handlers...)...)
}

// DELETE is a shortcut for Handle("DELETE", path, handlers).
func (r *Router) DELETE(path string, handlers ...HandlerFunc) {
	r.Engine.DELETE(path, wrapHandlers(handlers...)...)
}

// GET is a shortcut for Handle("GET", path, handlers).
func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.Engine.GET(path, wrapHandlers(handlers...)...)
}

// HEAD is a shortcut for Handle("HEAD", path, handlers).
func (r *Router) HEAD(path string, handlers ...HandlerFunc) {
	r.Engine.HEAD(path, wrapHandlers(handlers...)...)
}

// OPTIONS is a shortcut for Handle("OPTIONS", path, handlers).
func (r *Router) OPTIONS(path string, handlers ...HandlerFunc) {
	r.Engine.OPTIONS(path, wrapHandlers(handlers...)...)
}

// PATCH is a shortcut for Handle("PATCH", path, handlers).
func (r *Router) PATCH(path string, handlers ...HandlerFunc) {
	r.Engine.PATCH(path, wrapHandlers(handlers...)...)
}

// POST is a shortcut for Handle("POST", path, handlers).
func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.Engine.POST(path, wrapHandlers(handlers...)...)
}

// PUT is a shortcut for Handle("PUT", path, handlers).
func (r *Router) PUT(path string, handlers ...HandlerFunc) {
	r.Engine.PUT(path, wrapHandlers(handlers...)...)
}

// Use attaches a global middleware to the router.
func (r *Router) Use(handlers ...HandlerFunc) {
	r.Engine.Use(wrapHandlers(handlers...)...)
}

// Handle registers a new request handle with the method, given path and middleware.
func (rg *RouteGroup) Handle(method, path string, handlers ...HandlerFunc) {
	rg.RouterGroup.Handle(method, path, wrapHandlers(handlers...)...)
}

// Any registers a route that matches all the HTTP methods, i.e. GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE,
// CONNECT, TRACE.
func (rg *RouteGroup) Any(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.Any(path, wrapHandlers(handlers...)...)
}

// DELETE is a shortcut for Handle("DELETE", path, handlers).
func (rg *RouteGroup) DELETE(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.DELETE(path, wrapHandlers(handlers...)...)
}

// GET is a shortcut for Handle("GET", path, handlers).
func (rg *RouteGroup) GET(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.GET(path, wrapHandlers(handlers...)...)
}

// HEAD is a shortcut for Handle("HEAD", path, handlers).
func (rg *RouteGroup) HEAD(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.HEAD(path, wrapHandlers(handlers...)...)
}

// OPTIONS is a shortcut for Handle("OPTIONS", path, handlers).
func (rg *RouteGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.OPTIONS(path, wrapHandlers(handlers...)...)
}

// PATCH is a shortcut for Handle("PATCH", path, handlers).
func (rg *RouteGroup) PATCH(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.PATCH(path, wrapHandlers(handlers...)...)
}

// POST is a shortcut for Handle("POST", path, handlers).
func (rg *RouteGroup) POST(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.POST(path, wrapHandlers(handlers...)...)
}

// PUT is a shortcut for Handle("PUT", path, handlers).
func (rg *RouteGroup) PUT(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.PUT(path, wrapHandlers(handlers...)...)
}

// Use attaches a global middleware to the router.
func (rg *RouteGroup) Use(handlers ...HandlerFunc) {
	rg.RouterGroup.Use(wrapHandlers(handlers...)...)
}

func wrapHandler(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(&Context{
			Context: c,
		})
	}
}

func wrapHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	ginHandlers := []gin.HandlerFunc{}

	if len(handlers) > 0 {
		for _, handler := range handlers {
			ginHandlers = append(ginHandlers, wrapHandler(handler))
		}
	}

	return ginHandlers
}
