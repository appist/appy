package http

import (
	"html/template"
	"net/http"

	"appist/appy/middleware"
	"appist/appy/support"
	atpl "appist/appy/template"

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
	HTTP         *http.Server
	HTTPS        *http.Server
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewServer returns a ServerT instance.
func NewServer(config *support.ConfigT) *ServerT {
	renderer := multitemplate.NewRenderer()
	router := newRouter()
	router.HTMLRender = renderer
	h, hs := newServers(config, router)

	// Initialize the error templates.
	renderer.AddFromString("error/404", atpl.ErrorTpl404())
	renderer.AddFromString("error/500", atpl.ErrorTpl500())

	return &ServerT{
		htmlRenderer: renderer,
		router:       router,
		HTTP:         h,
		HTTPS:        hs,
	}
}

// GetAllRoutes returns all the routes including the ones in middlewares.
func (s *ServerT) GetAllRoutes() []gin.RouteInfo {
	routes := s.Routes()

	if support.Config.HTTPHealthCheckURL != "" {
		routes = append(routes, gin.RouteInfo{
			Method:      "GET",
			Path:        support.Config.HTTPHealthCheckURL,
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

func newRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.CSRF(support.Config))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.RealIP())
	r.Use(middleware.SessionManager(support.Config))
	r.Use(middleware.HealthCheck(support.Config.HTTPHealthCheckURL))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(support.Config)))
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

func newServers(config *support.ConfigT, router *gin.Engine) (*http.Server, *http.Server) {
	h := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	h.ErrorLog = zap.NewStdLog(support.Logger.Desugar())

	hs := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPSSLPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	hs.ErrorLog = zap.NewStdLog(support.Logger.Desugar())

	return h, hs
}
