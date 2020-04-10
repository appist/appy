package pack

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	gqlLRU "github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
	"github.com/gin-contrib/multitemplate"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

type (
	// Server processes the HTTP requests.
	Server struct {
		asset        *support.Asset
		config       *support.Config
		http         *http.Server
		https        *http.Server
		logger       *support.Logger
		middleware   []HandlerFunc
		router       *Router
		spaResources []*spaResource
	}

	spaResource struct {
		fs         http.FileSystem
		fileServer http.Handler
		prefix     string
	}
)

// NewServer initializes Server instance.
func NewServer(asset *support.Asset, config *support.Config, logger *support.Logger) *Server {
	router := newRouter()

	hs := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	hs.ErrorLog = zap.NewStdLog(logger.Desugar())

	hss := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPSSLPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	hss.ErrorLog = zap.NewStdLog(logger.Desugar())

	server := &Server{
		asset:        asset,
		config:       config,
		http:         hs,
		https:        hss,
		logger:       logger,
		middleware:   []HandlerFunc{},
		router:       router,
		spaResources: []*spaResource{},
	}

	return server
}

// BasePath returns the base path.
func (s *Server) BasePath() string {
	return s.router.BasePath()
}

// Config returns the server's configuration.
func (s *Server) Config() *support.Config {
	return s.config
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

// Info returns the server info.
func (s *Server) Info() []string {
	configPath := s.config.Path()
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			support.VERSION, runtime.Version(), support.Build, s.config.AppyEnv, configPath,
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.config.HTTPPort)

	if s.config.HTTPSSLEnabled {
		host += fmt.Sprintf(", https://%s:%s", hosts[0], s.config.HTTPSSLPort)
	}

	return append(lines, fmt.Sprintf("* Listening on %s", host))
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

// Router returns the router instance.
func (s *Server) Router() *Router {
	return s.router
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

// ServeHTTP conforms to the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// ServeNoRoute handles custom 404 not found error.
func (s *Server) ServeNoRoute() {
	// TODO: allow custom 404 page with translations.
	s.router.NoRoute(CSRFSkipCheck(), func(c *Context) {
		c.defaultHTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

// ServeSPA serves the SPA at the specified prefix path.
func (s *Server) ServeSPA(prefix string, fs http.FileSystem) {
	s.router.Use(mdwSPA(s, prefix, fs))
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

	queryCacheSize := 1000
	if s.Config().GQLQueryCacheSize > 0 {
		queryCacheSize = s.Config().GQLQueryCacheSize
	}
	gqlServer.SetQueryCache(gqlLRU.New(queryCacheSize))
	gqlServer.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		// Refer to https://gqlgen.com/reference/errors/#the-error-presenter for custom error handling.
		return graphql.DefaultErrorPresenter(ctx, err)
	})
	gqlServer.SetRecoverFunc(func(c context.Context, err interface{}) error {
		// TODO: Implement error alert.
		return err.(error)
	})

	gqlServer.Use(extension.Introspection{})

	APQCacheSize := 100
	if s.Config().GQLAPQCacheSize > 0 {
		APQCacheSize = s.Config().GQLAPQCacheSize
	}
	gqlServer.Use(extension.AutomaticPersistedQuery{
		Cache: gqlLRU.New(APQCacheSize),
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
			c.Data(http.StatusOK, "text/html; charset=utf-8", gqlPlaygroundTpl(path, c))
		})
	}
}

// SetupMailerPreview sets up the mailer preview for debugging purpose.
func (s *Server) SetupMailerPreview(m *mailer.Engine, i18n *support.I18n) {
	if support.IsReleaseBuild() {
		return
	}

	s.router.HTMLRender.(multitemplate.Renderer).AddFromString("mailer/preview", mailerPreviewTpl())

	// Serve the preview listing page.
	s.GET(s.config.MailerPreviewPath, func(c *Context) {
		name := c.DefaultQuery("name", "")
		if name == "" && len(m.Previews()) > 0 {
			for _, preview := range m.Previews() {
				name = preview.Template
				break
			}
		}

		locale := c.DefaultQuery("locale", s.config.I18nDefaultLocale)
		preview := &mailer.Mail{}

		if name != "" {
			preview = m.Previews()[name]
			preview.Locale = locale

			subject := i18n.T(preview.Subject, preview.Locale)
			if subject != "" {
				preview.Subject = subject
			}
		}

		c.defaultHTML(http.StatusOK, "mailer/preview", H{
			"path":          s.config.MailerPreviewPath,
			"previews":      m.Previews(),
			"title":         "Mailer Preview",
			"name":          name,
			"ext":           c.DefaultQuery("ext", "html"),
			"locale":        locale,
			"locales":       i18n.Locales(),
			"mail":          preview,
			"liveReloadTpl": template.HTML(liveReloadTpl(c.Request.Host, c.Request.TLS)),
		})
	})

	// Serve the preview content.
	s.GET(s.config.MailerPreviewPath+"/preview", func(c *Context) {
		name := c.Query("name")
		preview, exists := m.Previews()[name]

		if !exists {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var (
			contentType string
			content     []byte
		)

		preview.Locale = c.DefaultQuery("locale", s.config.I18nDefaultLocale)
		ext := c.DefaultQuery("ext", "html")
		switch ext {
		case "html":
			contentType = "text/html"
			email, err := m.ComposeEmail(preview)
			if err != nil {
				c.Logger().Error(err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			content = email.HTML
		case "txt":
			contentType = "text/plain"
			email, err := m.ComposeEmail(preview)
			if err != nil {
				c.Logger().Error(err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			content = email.Text
		}

		c.Writer.Header().Del(http.CanonicalHeaderKey("x-frame-options"))
		c.Data(http.StatusOK, contentType, content)
	})
}

// TestHTTPRequest provides a simple way to fire HTTP request to the server.
func (s *Server) TestHTTPRequest(method, path string, header H, body io.Reader) *ResponseRecorder {
	w := NewResponseRecorder()
	req, _ := http.NewRequest(method, path, body)

	for key, val := range header {
		req.Header.Add(key, val.(string))
	}

	s.ServeHTTP(w, req)
	return w
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

func (s *Server) isSSRPath(path string) bool {
	for _, route := range s.Routes() {
		if filepath.Ext(path) == "" && strings.Contains(path, route.Path) {
			return true
		}
	}

	return false
}

// ResponseRecorder is an implementation of http.ResponseWriter that records its mutations for later inspection in tests.
type ResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
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

// Close close the notify channel.
func (r *ResponseRecorder) Close() {
	r.closeChannel <- true
}
