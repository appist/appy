package http

import (
	"html/template"
	"net/http"

	"appist/appy/middleware"
	"appist/appy/support"
	at "appist/appy/template"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServerT is a high level HTTP server that provides more functionalities to serve HTTP requests.
type ServerT struct {
	assets       http.FileSystem
	funcMap      template.FuncMap
	htmlRenderer multitemplate.Renderer
	router       *gin.Engine
	Config       *support.ConfigT
	HTTP         *http.Server
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewServer returns a ServerT instance.
func NewServer(config *support.ConfigT) *ServerT {
	renderer := multitemplate.NewRenderer()
	router := newRouter(config)
	router.HTMLRender = renderer
	server := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	server.ErrorLog = zap.NewStdLog(support.Logger.Desugar())

	if config.HTTPSSLEnabled == true {
		server.Addr = config.HTTPHost + ":" + config.HTTPSSLPort
	}

	// Initialize the error templates.
	renderer.AddFromString("error/404", at.ErrorTpl404())
	renderer.AddFromString("error/500", at.ErrorTpl500())

	return &ServerT{
		htmlRenderer: renderer,
		router:       router,
		Config:       config,
		HTTP:         server,
	}
}

// GetAllRoutes returns all the routes including the ones in middlewares.
func (s *ServerT) GetAllRoutes() []gin.RouteInfo {
	routes := s.Routes()

	if s.Config.HTTPHealthCheckURL != "" {
		routes = append(routes, gin.RouteInfo{
			Method:      "GET",
			Path:        s.Config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// SecureJSONPrefix sets the prefix used in RenderSecureJSON.
func (s *ServerT) SecureJSONPrefix(prefix string) *gin.Engine {
	return s.router.SecureJsonPrefix(prefix)
}

// SetupAssets sets up the static assets.
func (s *ServerT) SetupAssets(assets http.FileSystem) {
	s.assets = assets
}

// SetFuncMap sets the view template helpers.
func (s *ServerT) SetFuncMap(fm template.FuncMap) {
	nfm := template.FuncMap{}

	for n := range fm {
		nfm[n] = fm[n]
	}

	s.funcMap = nfm
}

func newRouter(config *support.ConfigT) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CSRF(config))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(config))
	r.Use(middleware.RealIP())
	r.Use(middleware.SessionManager(config))
	r.Use(middleware.HealthCheck(config.HTTPHealthCheckURL))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(config)))
	r.Use(middleware.Recovery())

	return r
}

func newSecureConfig(config *support.ConfigT) secure.Config {
	return secure.Config{
		IsDevelopment:           false,
		AllowedHosts:            config.HTTPAllowedHosts,
		SSLRedirect:             config.HTTPSSLRedirect,
		SSLTemporaryRedirect:    config.HTTPSSLTemporaryRedirect,
		SSLHost:                 config.HTTPSSLHost,
		STSSeconds:              config.HTTPSTSSeconds,
		STSIncludeSubdomains:    config.HTTPSTSIncludeSubdomains,
		FrameDeny:               config.HTTPFrameDeny,
		CustomFrameOptionsValue: config.HTTPCustomFrameOptionsValue,
		ContentTypeNosniff:      config.HTTPContentTypeNosniff,
		BrowserXssFilter:        config.HTTPBrowserXSSFilter,
		ContentSecurityPolicy:   config.HTTPContentSecurityPolicy,
		ReferrerPolicy:          config.HTTPReferrerPolicy,
		IENoOpen:                config.HTTPIENoOpen,
		SSLProxyHeaders:         config.HTTPSSLProxyHeaders,
	}
}
