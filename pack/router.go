package pack

import (
	"net/http"
	"reflect"
	"runtime"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

type (
	// HandlerFunc defines the handler used by middleware as return value.
	HandlerFunc func(*Context)

	// HandlersChain defines the HandlerFunc array.
	HandlersChain []HandlerFunc

	// Route represents a request route's specification which contains method
	// and path and its handler.
	Route struct {
		Method      string
		Path        string
		Handler     string
		HandlerFunc HandlerFunc
	}

	// Routes defines the array of Route.
	Routes []Route

	// Router maintains the routing logic.
	Router struct {
		*gin.Engine
		internalRoutes map[string]Route
	}
)

var (
	anyMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodTrace,
	}
)

func newRouter() *Router {
	r := &Router{
		gin.New(),
		map[string]Route{},
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

// Group creates a new route group. You should add all the routes that have
// common middlewares or the same path prefix.
func (r *Router) Group(path string, handlers ...HandlerFunc) *RouteGroup {
	group := r.RouterGroup.Group(path, wrapHandlers(handlers...)...)

	return &RouteGroup{
		group,
		r.internalRoutes,
	}
}

// Handle registers a new request handle with the method, given path and
// middleware.
func (r *Router) Handle(method, path string, handlers ...HandlerFunc) {
	r.Engine.Handle(method, path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, method, path, handlers...)
}

// NoRoute adds handlers for NoRoute which returns a 404 code by default.
func (r *Router) NoRoute(handlers ...HandlerFunc) {
	r.Engine.NoRoute(wrapHandlers(handlers...)...)
}

// Any registers a route that matches all the HTTP methods, i.e. GET, POST,
// PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (r *Router) Any(path string, handlers ...HandlerFunc) {
	r.Engine.Any(path, wrapHandlers(handlers...)...)

	for _, method := range anyMethods {
		appendInternalRoute(r.internalRoutes, method, path, handlers...)
	}
}

// DELETE is a shortcut for Handle("DELETE", path, handlers).
func (r *Router) DELETE(path string, handlers ...HandlerFunc) {
	r.Engine.DELETE(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "DELETE", path, handlers...)
}

// GET is a shortcut for Handle("GET", path, handlers).
func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.Engine.GET(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "GET", path, handlers...)
}

// HEAD is a shortcut for Handle("HEAD", path, handlers).
func (r *Router) HEAD(path string, handlers ...HandlerFunc) {
	r.Engine.HEAD(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "HEAD", path, handlers...)
}

// OPTIONS is a shortcut for Handle("OPTIONS", path, handlers).
func (r *Router) OPTIONS(path string, handlers ...HandlerFunc) {
	r.Engine.OPTIONS(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "OPTIONS", path, handlers...)
}

// PATCH is a shortcut for Handle("PATCH", path, handlers).
func (r *Router) PATCH(path string, handlers ...HandlerFunc) {
	r.Engine.PATCH(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "PATCH", path, handlers...)
}

// POST is a shortcut for Handle("POST", path, handlers).
func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.Engine.POST(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "POST", path, handlers...)
}

// PUT is a shortcut for Handle("PUT", path, handlers).
func (r *Router) PUT(path string, handlers ...HandlerFunc) {
	r.Engine.PUT(path, wrapHandlers(handlers...)...)
	appendInternalRoute(r.internalRoutes, "PUT", path, handlers...)
}

// Use attaches a global middleware to the router.
func (r *Router) Use(handlers ...HandlerFunc) {
	r.Engine.Use(wrapHandlers(handlers...)...)
}

func (r *Router) routes() []Route {
	routes := []Route{}

	for _, route := range r.Routes() {
		routes = append(routes, Route{
			Method:      route.Method,
			Path:        route.Path,
			Handler:     r.internalRoutes[route.Method+" "+route.Path].Handler,
			HandlerFunc: r.internalRoutes[route.Method+" "+route.Path].HandlerFunc,
		})
	}

	return routes
}

// RouteGroup is associated with a prefix and an array of handlers.
type RouteGroup struct {
	*gin.RouterGroup
	internalRoutes map[string]Route
}

// Group creates a new route group. You should add all the routes that have
// common middlewares or the same path prefix.
func (rg *RouteGroup) Group(path string, handlers ...HandlerFunc) *RouteGroup {
	group := rg.RouterGroup.Group(path, wrapHandlers(handlers...)...)

	return &RouteGroup{
		group,
		rg.internalRoutes,
	}
}

// Handle registers a new request handle with the method, given path and
// middleware.
func (rg *RouteGroup) Handle(method, path string, handlers ...HandlerFunc) {
	rg.RouterGroup.Handle(method, path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, method, rg.RouterGroup.BasePath()+path, handlers...)
}

// Any registers a route that matches all the HTTP methods, i.e. GET, POST,
// PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (rg *RouteGroup) Any(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.Any(path, wrapHandlers(handlers...)...)

	for _, method := range anyMethods {
		appendInternalRoute(rg.internalRoutes, method, rg.RouterGroup.BasePath()+path, handlers...)
	}
}

// DELETE is a shortcut for Handle("DELETE", path, handlers).
func (rg *RouteGroup) DELETE(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.DELETE(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "DELETE", rg.RouterGroup.BasePath()+path, handlers...)
}

// GET is a shortcut for Handle("GET", path, handlers).
func (rg *RouteGroup) GET(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.GET(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "GET", rg.RouterGroup.BasePath()+path, handlers...)
}

// HEAD is a shortcut for Handle("HEAD", path, handlers).
func (rg *RouteGroup) HEAD(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.HEAD(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "HEAD", rg.RouterGroup.BasePath()+path, handlers...)
}

// OPTIONS is a shortcut for Handle("OPTIONS", path, handlers).
func (rg *RouteGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.OPTIONS(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "OPTIONS", rg.RouterGroup.BasePath()+path, handlers...)
}

// PATCH is a shortcut for Handle("PATCH", path, handlers).
func (rg *RouteGroup) PATCH(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.PATCH(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "PATCH", rg.RouterGroup.BasePath()+path, handlers...)
}

// POST is a shortcut for Handle("POST", path, handlers).
func (rg *RouteGroup) POST(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.POST(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "POST", rg.RouterGroup.BasePath()+path, handlers...)
}

// PUT is a shortcut for Handle("PUT", path, handlers).
func (rg *RouteGroup) PUT(path string, handlers ...HandlerFunc) {
	rg.RouterGroup.PUT(path, wrapHandlers(handlers...)...)
	appendInternalRoute(rg.internalRoutes, "PUT", rg.RouterGroup.BasePath()+path, handlers...)
}

// Use attaches a global middleware to the router.
func (rg *RouteGroup) Use(handlers ...HandlerFunc) {
	rg.RouterGroup.Use(wrapHandlers(handlers...)...)
}

func appendInternalRoute(internalRoutes map[string]Route, method, path string, handlers ...HandlerFunc) {
	var (
		lastHandlerName string
		lastHandler     HandlerFunc
	)

	if len(handlers) > 0 {
		lastHandler = handlers[len(handlers)-1]
		lastHandlerName = nameOfFunction(lastHandler)
	}

	internalRoutes[method+" "+path] = Route{
		Method:      method,
		Path:        path,
		Handler:     lastHandlerName,
		HandlerFunc: lastHandler,
	}
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

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
