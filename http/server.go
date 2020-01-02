package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	gqlLRU "github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/appist/appy/support"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/gqlerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	// Server serves the HTTP requests.
	Server struct {
		assets       *support.Assets
		config       *support.Config
		grpc         *grpc.Server
		http         *http.Server
		https        *http.Server
		logger       *support.Logger
		middleware   []HandlerFunc
		router       *Router
		spaResources []*spaResource
	}

	spaResource struct {
		assets     http.FileSystem
		fileServer http.Handler
		prefix     string
	}

	// ResponseRecorder is an implementation of http.ResponseWriter that records its mutations for later inspection in tests.
	ResponseRecorder struct {
		*httptest.ResponseRecorder
		closeChannel chan bool
	}
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewResponseRecorder returns an initialized ResponseRecorder for testing purpose.
func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

// CloseNotify implements http.CloseNotifier.
func (r *ResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *ResponseRecorder) closeClient() {
	r.closeChannel <- true
}

// NewServer initializes Server instance.
func NewServer(assets *support.Assets, config *support.Config, logger *support.Logger) *Server {
	router := newRouter()

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
		assets:     assets,
		config:     config,
		grpc:       grpc.NewServer(),
		http:       httpServer,
		https:      httpsServer,
		logger:     logger,
		middleware: []HandlerFunc{},
		router:     router,
	}
}

// Config returns the server's configuration.
func (s *Server) Config() *support.Config {
	return s.config
}

// SetupGraphQL sets up the GraphQL stack.
func (s *Server) SetupGraphQL(path string, es graphql.ExecutableSchema, exts []graphql.HandlerExtension) {
	gqlServer := gqlHandler.New(es)
	gqlServer.AddTransport(transport.Websocket{
		KeepAlivePingInterval: s.Config().GQLWebsocketKeepAliveDuration,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	gqlServer.AddTransport(transport.Options{})
	gqlServer.AddTransport(transport.GET{})
	gqlServer.AddTransport(transport.POST{})
	gqlServer.AddTransport(transport.MultipartForm{
		MaxMemory:     s.Config().GQLMultipartMaxMemory,
		MaxUploadSize: s.Config().GQLMultipartMaxUploadSize,
	})

	gqlServer.SetQueryCache(gqlLRU.New(s.Config().GQLQueryCacheSize))
	gqlServer.SetErrorPresenter(func(c context.Context, err error) *gqlerror.Error {
		// Refer to https://gqlgen.com/reference/errors/#the-error-presenter for custom error handling.
		return graphql.DefaultErrorPresenter(c, err)
	})
	gqlServer.SetRecoverFunc(func(c context.Context, err interface{}) error {
		// TODO: Implement error alert.
		return err.(error)
	})

	gqlServer.Use(extension.Introspection{})
	gqlServer.Use(extension.AutomaticPersistedQuery{
		Cache: gqlLRU.New(s.Config().GQLAPQCacheSize),
	})
	gqlServer.Use(extension.FixedComplexityLimit(s.Config().GQLComplexityLimit))
	gqlServer.Use(apollotracing.Tracer{})

	for _, ext := range exts {
		gqlServer.Use(ext)
	}

	s.router.Any(path, func(c *Context) {
		gqlServer.ServeHTTP(c.Writer, c.Request)
	})

	if s.config.GQLPlaygroundEnabled && s.config.GQLPlaygroundPath != "" {
		s.router.GET(s.config.GQLPlaygroundPath, CSRFSkipCheck(), func(c *Context) {
			c.Data(http.StatusOK, "text/html", gqlPlaygroundTpl(path, c))
		})
	}
}

func gqlPlaygroundTpl(path string, c *Context) []byte {
	return []byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset=utf-8/>
	<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
	<title>GraphQL Playground</title>
	<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
	<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
	<script src="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
	<div id="root">
	<style>
		body { background-color: rgb(23, 42, 58); font-family: Open Sans, sans-serif; height: 90vh; }
		#root { height: 100%; width: 100%; display: flex; align-items: center; justify-content: center; }
		.loading { font-size: 32px; font-weight: 200; color: rgba(255, 255, 255, .6); margin-left: 20px; }
		img { width: 78px; height: 78px; }
		.title { font-weight: 400; }
	</style>
	<img src='//cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
	<div class="loading"> Loading
		<span class="title">GraphQL Playground</span>
	</div>
	</div>
	<script>
		function getCookie(name) {
			var v = document.cookie.match('(^|;) ?' + name + '=([^;]*)(;|$)');
			return v ? v[2] : null;
		}
		window.addEventListener('load', function (event) {
			GraphQLPlayground.init(document.getElementById('root'), {
				endpoint: '` + path + `',
				subscriptionEndpoint: '` + path + `',
				headers: {
					'X-CSRF-Token': unescape(getCookie("` + csrfTemplateFieldName(c) + `"))
				},
				settings: {
					'request.credentials': 'include',
					'schema.polling.interval': 5000
				}
			})
		})
	</script>
</body>
</html>
`)
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

// HTMLRenderer returns the HTML renderer instance.
func (s *Server) HTMLRenderer() multitemplate.Renderer {
	return s.router.HTMLRender.(multitemplate.Renderer)
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

	if s.config.HTTPSSLEnabled {
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

func (s *Server) spaResource(path string) *spaResource {
	var (
		resource, rootResource *spaResource
	)

	for _, res := range s.spaResources {
		if res.prefix == "/" {
			rootResource = res
		}

		if res.prefix != "/" && strings.HasPrefix(path, res.prefix) {
			resource = res
		}
	}

	if resource == nil {
		resource = rootResource
	}

	return resource
}

// ServeSPA serves the SPA at the specified prefix path.
func (s *Server) ServeSPA(prefix string, assets http.FileSystem) {
	s.router.Use(SPA(s, prefix, assets))
}

// ServeNoRoute handles custom 404 not found error.
func (s *Server) ServeNoRoute() {
	// TODO: allow custom 404 page with translations.
	s.router.NoRoute(CSRFSkipCheck(), func(c *Context) {
		c.DefaultHTML(http.StatusNotFound, "error/404", support.H{
			"title": "404 Page Not Found",
		})
	})
}

// TestHTTPRequest provides a simple way to fire HTTP request to the server.
func (s *Server) TestHTTPRequest(method, path string, header support.H, body io.Reader) *ResponseRecorder {
	w := NewResponseRecorder()
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
	s.middleware = append(s.middleware, handlers...)
	s.router.Use(handlers...)
}

// Middleware returns the global middleware list.
func (s *Server) Middleware() []HandlerFunc {
	return s.middleware
}

func (s *Server) isSSRPath(path string) bool {
	for _, route := range s.Routes() {
		if strings.Contains(path, route.Path) {
			return true
		}
	}

	return false
}
