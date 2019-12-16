package http

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"

	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	// Server serves the HTTP requests.
	Server struct {
		assetsMngr *support.AssetsMngr
		config     *support.Config
		grpc       *grpc.Server
		http       *http.Server
		https      *http.Server
		logger     *support.Logger
		router     *Router
	}
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewServer initializes Server instance.
func NewServer(assetsMngr *support.AssetsMngr, config *support.Config, logger *support.Logger) *Server {
	router := newRouter(config, logger)

	httpServer := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	httpServer.ErrorLog = zap.NewStdLog(logger.Desugar())

	httpsServer := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPSSLPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	httpsServer.ErrorLog = zap.NewStdLog(logger.Desugar())

	return &Server{
		assetsMngr: assetsMngr,
		config:     config,
		grpc:       grpc.NewServer(),
		http:       httpServer,
		https:      httpsServer,
		logger:     logger,
		router:     router,
	}
}

// Config returns the server's configuration.
func (s *Server) Config() *support.Config {
	return s.config
}

// GRPC returns the GRPC server instance.
func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}

// HTTP returns the HTTP server instance.
func (s *Server) HTTP() *http.Server {
	return s.http
}

// HTTPS returns the HTTPS server instance.
func (s *Server) HTTPS() *http.Server {
	return s.https
}

// Hosts returns the server hosts list.
func (s *Server) Hosts() ([]string, error) {
	var hosts = []string{}

	if s.config.HTTPHost != "" && !support.ArrayContains(hosts, s.config.HTTPHost) {
		hosts = append(hosts, s.config.HTTPHost)
	}

	addresses, _ := net.InterfaceAddrs()
	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			host := ipnet.IP.To4()
			if host != nil {
				hosts = append(hosts, host.String())
			}
		}
	}

	return hosts, nil
}

// IsSSLCertExisted checks if `./tmp/ssl` exists and contains the locally trusted SSL certificates.
func (s *Server) IsSSLCertExisted() bool {
	_, certErr := os.Stat(s.config.HTTPSSLCertPath + "/cert.pem")
	_, keyErr := os.Stat(s.config.HTTPSSLCertPath + "/key.pem")

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return false
	}

	return true
}

// Info returns the server info.
func (s *Server) Info() []string {
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			support.VERSION, runtime.Version(), support.Build, s.config.AppyEnv, s.config.Path(),
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.config.HTTPPort)

	if s.config.HTTPSSLEnabled == true {
		host += fmt.Sprintf(", https://%s:%s", hosts[0], s.config.HTTPSSLPort)
	}

	return append(lines, fmt.Sprintf("* Listening on %s", host))
}

// Routes returns all the routes including those in middlewares.
func (s *Server) Routes() []Route {
	routes := s.router.routes()

	if s.config.HTTPHealthCheckURL != "" {
		routes = append(routes, Route{
			Method:      "GET",
			Path:        s.config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// TestHTTPRequest provides a simple way to fire HTTP request to the server.
func (s *Server) TestHTTPRequest(method, path string, header H, body io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, body)

	for key, val := range header {
		req.Header.Add(key, val.(string))
	}

	s.ServeHTTP(w, req)
	return w
}

// ServeHTTP conforms to the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// BasePath returns the base path.
func (s *Server) BasePath() string {
	return s.router.BasePath()
}

// Group creates a new route group. You should add all the routes that have common middlewares or the same path prefix.
func (s *Server) Group(path string, handlers ...HandlerFunc) *RouteGroup {
	return s.router.Group(path, handlers...)
}

// Handle registers a new request handle with the method, given path and middleware.
func (s *Server) Handle(method, path string, handlers ...HandlerFunc) {
	s.router.Handle(method, path, handlers...)
}

// Any registers a route that matches all the HTTP methods, i.e. GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE,
// CONNECT, TRACE.
func (s *Server) Any(path string, handlers ...HandlerFunc) {
	s.router.Any(path, handlers...)
}

// DELETE is a shortcut for Handle("DELETE", path, handlers).
func (s *Server) DELETE(path string, handlers ...HandlerFunc) {
	s.router.DELETE(path, handlers...)
}

// GET is a shortcut for Handle("GET", path, handlers).
func (s *Server) GET(path string, handlers ...HandlerFunc) {
	s.router.GET(path, handlers...)
}

// HEAD is a shortcut for Handle("HEAD", path, handlers).
func (s *Server) HEAD(path string, handlers ...HandlerFunc) {
	s.router.HEAD(path, handlers...)
}

// OPTIONS is a shortcut for Handle("OPTIONS", path, handlers).
func (s *Server) OPTIONS(path string, handlers ...HandlerFunc) {
	s.router.OPTIONS(path, handlers...)
}

// PATCH is a shortcut for Handle("PATCH", path, handlers).
func (s *Server) PATCH(path string, handlers ...HandlerFunc) {
	s.router.PATCH(path, handlers...)
}

// POST is a shortcut for Handle("POST", path, handlers).
func (s *Server) POST(path string, handlers ...HandlerFunc) {
	s.router.POST(path, handlers...)
}

// PUT is a shortcut for Handle("PUT", path, handlers).
func (s *Server) PUT(path string, handlers ...HandlerFunc) {
	s.router.PUT(path, handlers...)
}

// Use attaches a global middleware to the router.
func (s *Server) Use(handlers ...HandlerFunc) {
	s.router.Use(handlers...)
}

// Middleware returns the global middleware list.
func (s *Server) Middleware() gin.HandlersChain {
	return s.router.Handlers
}
