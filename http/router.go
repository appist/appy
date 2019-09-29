package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BasePath returns the base path of router group.
// For example, if v := router.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".
func (s *ServerT) BasePath() string {
	return s.router.BasePath()
}

// Group creates a new router group. You should add all the routes that have common middlewares or
// the same path prefix. For example, all the routes that use a common middleware for authorization
// could be grouped.
func (s *ServerT) Group(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.Group(p, handlers...)
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should
// be shared among different routes.
//
// This function is intended for bulk loading and to allow the usage of less frequently used, non-standardized
// or custom methods (e.g. for internal communication with a proxy).
func (s *ServerT) Handle(httpMethod, p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.Handle(httpMethod, p, handlers...)
}

// Use adds middleware to the group.
func (s *ServerT) Use(middlewares ...gin.HandlerFunc) gin.IRoutes {
	return s.router.Use(middlewares...)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (s *ServerT) Any(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.Any(p, handlers...)
}

// DELETE sets up a DELETE HTTP endpoint.
func (s *ServerT) DELETE(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.DELETE(p, handlers...)
}

// HEAD sets up a HEAD HTTP endpoint.
func (s *ServerT) HEAD(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.HEAD(p, handlers...)
}

// GET sets up a GET HTTP endpoint.
func (s *ServerT) GET(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.GET(p, handlers...)
}

// OPTIONS sets up a OPTIONS HTTP endpoint.
func (s *ServerT) OPTIONS(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.OPTIONS(p, handlers...)
}

// POST sets up a POST HTTP endpoint.
func (s *ServerT) POST(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.POST(p, handlers...)
}

// PATCH sets up a PATCH HTTP endpoint.
func (s *ServerT) PATCH(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.PATCH(p, handlers...)
}

// PUT sets up a PUT HTTP endpoint.
func (s *ServerT) PUT(p string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return s.router.PUT(p, handlers...)
}

// Middlewares returns a slice of registered middlewares for the router singleton.
func (s *ServerT) Middlewares() gin.HandlersChain {
	return s.router.Handlers
}

// Routes returns a slice of registered routes which includes http method, path and handler name.
func (s *ServerT) Routes() []gin.RouteInfo {
	return s.router.Routes()
}

// ServeHTTP conforms to the http.Handler interface.
func (s *ServerT) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// Static serves files from the given file system root. Internally a http.FileServer is used, therefore
// http.NotFound is used instead of the Router's NotFound handler. To use the operating system's file system
// implementation, use: router.Static("/static", "/var/www")
func (s *ServerT) Static(p string, dir string) gin.IRoutes {
	return s.router.Static(p, dir)
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
func (s *ServerT) StaticFS(p string, fs http.FileSystem) gin.IRoutes {
	return s.router.StaticFS(p, fs)
}

// StaticFile registers a single route in order to serve a single file of the local filesystem.
func (s *ServerT) StaticFile(p string, filepath string) gin.IRoutes {
	return s.router.StaticFile(p, filepath)
}
