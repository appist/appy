package core

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/appist/appy/support"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// H is a type alias to gin.H.
type H = gin.H

// HandlerFunc is a type alias to gin.HandlerFunc.
type HandlerFunc = gin.HandlerFunc

// Context is a type alias to gin.Context.
type Context = gin.Context

// Router is a type alias to gin.Engine.
type Router = gin.Engine

// RouterGroup is a type alias to gin.RouterGroup.
type RouterGroup = gin.RouterGroup

// RouteInfo is a type alias to gin.RouteInfo.
type RouteInfo = gin.RouteInfo

// Routes is a type alias to gin.IRoutes.
type Routes = gin.IRoutes

type csrResource struct {
	assets     http.FileSystem
	fileServer http.Handler
	prefix     string
}

// AppServer is the core that serves HTTP/GRPC requests.
type AppServer struct {
	assets       http.FileSystem
	Config       AppConfig
	csrResources []csrResource
	GRPC         *grpc.Server
	HTTP         *http.Server
	htmlRenderer multitemplate.Renderer
	i18nBundle   *i18n.Bundle
	Logger       *AppLogger
	Router       *Router
	CSRPaths     map[string]string
	SSRPaths     map[string]string
	viewHelper   template.FuncMap
}

var (
	reservedViewDirs = []string{"layouts", "shared"}
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func newServer(assets http.FileSystem, c AppConfig, l *AppLogger, vh template.FuncMap) AppServer {
	// Initialize the error templates.
	renderer := multitemplate.NewRenderer()
	renderer.AddFromString("error/404", errorTpl404())
	renderer.AddFromString("error/500", errorTpl500())
	renderer.AddFromString("default/welcome", welcomeTpl())

	r := newRouter(c, l)
	r.HTMLRender = renderer

	http := &http.Server{
		Addr:              c.HTTPHost + ":" + c.HTTPPort,
		Handler:           r,
		MaxHeaderBytes:    c.HTTPMaxHeaderBytes,
		ReadTimeout:       c.HTTPReadTimeout,
		ReadHeaderTimeout: c.HTTPReadHeaderTimeout,
		WriteTimeout:      c.HTTPWriteTimeout,
		IdleTimeout:       c.HTTPIdleTimeout,
	}
	http.ErrorLog = zap.NewStdLog(l.Desugar())

	if c.HTTPSSLEnabled == true {
		http.Addr = c.HTTPHost + ":" + c.HTTPSSLPort
	}

	return AppServer{
		assets:       assets,
		Config:       c,
		csrResources: []csrResource{},
		GRPC:         nil,
		HTTP:         http,
		htmlRenderer: renderer,
		Logger:       l,
		Router:       r,
		CSRPaths:     CSRPaths,
		SSRPaths:     SSRPaths,
		viewHelper:   vh,
	}
}

func newRouter(c AppConfig, l *AppLogger) *gin.Engine {
	r := gin.New()
	r.AppEngine = true
	r.HandleMethodNotAllowed = true

	r.Use(CSRF(c, l))
	r.Use(RequestID())
	r.Use(RequestLogger(c, l))
	r.Use(RealIP())
	r.Use(ResponseHeaderFilter())
	r.Use(SessionManager(c))
	r.Use(HealthCheck(c.HTTPHealthCheckURL))
	r.Use(Prerender(c, l))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(c)))
	r.Use(Recovery(l))

	return r
}

func newSecureConfig(c AppConfig) secure.Config {
	return secure.Config{
		IsDevelopment:           false,
		AllowedHosts:            c.HTTPAllowedHosts,
		SSLRedirect:             c.HTTPSSLRedirect,
		SSLTemporaryRedirect:    c.HTTPSSLTemporaryRedirect,
		SSLHost:                 c.HTTPSSLHost,
		STSSeconds:              c.HTTPSTSSeconds,
		STSIncludeSubdomains:    c.HTTPSTSIncludeSubdomains,
		FrameDeny:               c.HTTPFrameDeny,
		CustomFrameOptionsValue: c.HTTPCustomFrameOptionsValue,
		ContentTypeNosniff:      c.HTTPContentTypeNosniff,
		BrowserXssFilter:        c.HTTPBrowserXSSFilter,
		ContentSecurityPolicy:   c.HTTPContentSecurityPolicy,
		ReferrerPolicy:          c.HTTPReferrerPolicy,
		IENoOpen:                c.HTTPIENoOpen,
		SSLProxyHeaders:         c.HTTPSSLProxyHeaders,
	}
}

// AddDefaultWelcomePage adds the default welcome page for `/` route.
func (s AppServer) AddDefaultWelcomePage() {
	routes := s.Routes()
	rootDefined := false

	for _, route := range routes {
		if route.Path == "/" {
			rootDefined = true
			break
		}
	}

	csrRootDefined := false
	if s.assets != nil {
		if _, err := s.assets.Open("/index.html"); err == nil {
			csrRootDefined = true
		}
	}

	if !rootDefined && !csrRootDefined {
		s.Router.GET("/", func(c *Context) {
			c.HTML(200, "default/welcome", nil)
		})
	}
}

// IsSSLCertsExist checks if `./tmp/ssl` exists and contains the locally trusted SSL certificates.
func (s AppServer) IsSSLCertsExist() bool {
	_, certErr := os.Stat(s.Config.HTTPSSLCertPath + "/cert.pem")
	_, keyErr := os.Stat(s.Config.HTTPSSLCertPath + "/key.pem")

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return false
	}

	return true
}

// Hosts returns the server hosts list.
func (s AppServer) Hosts() ([]string, error) {
	var hosts = []string{s.Config.HTTPHost}

	if s.Config.HTTPHost != "localhost" {
		hosts = append(hosts, "localhost")
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

// Routes returns all the routes including those in middlewares.
func (s AppServer) Routes() []RouteInfo {
	routes := s.Router.Routes()

	if s.Config.HTTPHealthCheckURL != "" {
		routes = append(routes, RouteInfo{
			Method:      "GET",
			Path:        s.Config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// PrintInfo prints the server info.
func (s AppServer) PrintInfo() {
	configPath, _, _ := getConfigInfo(s.assets)
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* Version %s (%s), build: %s, environment: %s, config: %s",
			VERSION, runtime.Version(), Build, s.Config.AppyEnv, configPath,
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.Config.HTTPPort)

	if s.Config.HTTPSSLEnabled == true {
		host = fmt.Sprintf("https://%s:%s", hosts[0], s.Config.HTTPSSLPort)
	}

	lines = append(lines, fmt.Sprintf("* Listening on %s", host))

	for _, line := range lines {
		if Build == "debug" {
			fmt.Println(line)
		} else {
			s.Logger.Info(line)
		}
	}
}

// InitCSR setup the client-side rendering/routing with index.html fallback.
func (s AppServer) InitCSR() {
	// Setup CSR hosting at "/".
	s.Router.Use(s.serveCSR("/", s.assets))

	// Setup CSR hosting at "/tools".
	// s.Router.Use(s.serveCSR("/tools", tools.Assets))

	s.Router.NoRoute(CSRFSkipCheck(), func(ctx *Context) {
		request := ctx.Request

		if request.Method == "GET" && !staticExtRegex.MatchString(request.URL.Path) {
			ctx.Header("Cache-Control", "no-cache")
			resource := s.csrResource(request.URL.Path)

			if resource.assets != nil {
				file, _ := resource.assets.Open("/index.html")

				if file != nil {
					data, _ := ioutil.ReadAll(file)

					if data != nil {
						ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
						return
					}
				}
			}
		}

		ctx.HTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

func (s AppServer) csrResource(path string) csrResource {
	var (
		resource, rootResource csrResource
	)

	for _, res := range s.csrResources {
		if res.prefix == "/" {
			rootResource = res
		}

		if res.prefix != "/" && strings.HasPrefix(path, res.prefix) {
			resource = res
		}
	}

	if (csrResource{}) == resource {
		resource = rootResource
	}

	return resource
}

func (s *AppServer) serveCSR(prefix string, assets http.FileSystem) HandlerFunc {
	s.csrResources = append(s.csrResources, csrResource{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(ctx *Context) {
		request := ctx.Request

		// Serve from the assets FS if the URL path isn't matching any of the SSR paths.
		if !isSSRPath(s.Routes(), request.URL.Path) && !strings.HasPrefix(request.URL.Path, "/"+s.SSRPaths["root"]) {
			resource := s.csrResource(request.URL.Path)

			if resource.assets != nil {
				assetPath := strings.Replace(request.URL.Path, resource.prefix, "", 1)
				_, err := resource.assets.Open(assetPath)
				// Only serve the request from assets if the file is in the assets filesystem.
				if err == nil {
					resource.fileServer.ServeHTTP(ctx.Writer, request)
					ctx.Abort()
				}
			}
		}
	}
}

// InitSSR initiates the SSR setup.
func (s *AppServer) InitSSR() error {
	if err := s.initSSRLocale(); err != nil {
		return err
	}

	if err := s.initSSRView(); err != nil {
		return err
	}

	return nil
}

func (s *AppServer) initSSRLocale() error {
	s.i18nBundle = i18n.NewBundle(language.English)
	s.i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var (
		localeFiles []os.FileInfo
		data        []byte
		err         error
	)
	localeDir := s.SSRPaths["locale"]

	// Try getting all the locale files from `app/locales`, but fallback to `assets` http.FileSystem.
	if Build == "debug" {
		localeFiles, err = ioutil.ReadDir(localeDir)

		if err != nil {
			return err
		}
	} else {
		localeDir = "/" + s.SSRPaths["root"] + "/" + localeDir
		file, err := s.assets.Open(localeDir)

		if err != nil {
			return err
		}

		localeFiles, _ = file.Readdir(-1)
	}

	for _, localeFile := range localeFiles {
		localeFn := localeFile.Name()

		if Build == "debug" {
			data, _ = ioutil.ReadFile(localeDir + "/" + localeFn)
		} else {
			file, err := s.assets.Open(localeDir + "/" + localeFn)
			if err != nil {
				return err
			}

			data, _ = ioutil.ReadAll(file)
		}

		s.i18nBundle.MustParseMessageFileBytes(data, localeFn)
	}

	s.Router.Use(I18n(s.i18nBundle))
	return nil
}

func (s *AppServer) initSSRView() error {
	var (
		fis []os.FileInfo
		err error
	)

	viewDir := s.SSRPaths["view"]

	// We will always read from local file system when it's debug build. Otherwise, read from the bind assets.
	if Build == "debug" {
		if fis, err = ioutil.ReadDir(viewDir); err != nil {
			return err
		}
	} else {
		viewDir = "/" + s.SSRPaths["root"] + "/" + viewDir

		var file http.File
		if file, err = s.assets.Open(viewDir); err != nil {
			return err
		}

		fis, err = file.Readdir(-1)
	}

	commonTpls := []string{}
	for _, fi := range fis {
		// We should only see directories in `app/views`.
		if fi.IsDir() == false {
			continue
		}

		if support.ArrayContains(reservedViewDirs, fi.Name()) == true {
			tpls, err := getCommonTemplates(s.assets, Build, viewDir+"/"+fi.Name())
			if err != nil {
				return err
			}

			commonTpls = append(commonTpls, tpls...)
		}
	}

	for _, fi := range fis {
		if fi.IsDir() == false || support.ArrayContains(reservedViewDirs, fi.Name()) == true {
			continue
		}

		var fileInfos []os.FileInfo
		targetDir := viewDir + "/" + fi.Name()
		if Build == "debug" {
			fileInfos, _ = ioutil.ReadDir(targetDir)
		} else {
			var file http.File
			if file, err = s.assets.Open(targetDir); err != nil {
				return err
			}

			fileInfos, _ = file.Readdir(-1)
		}

		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() == true {
				continue
			}

			viewName := fi.Name() + "/" + fileInfo.Name()
			targetFn := targetDir + "/" + fileInfo.Name()
			data, err := getTemplateContent(s.assets, Build, targetFn)
			if err != nil {
				return err
			}

			commonTplsCopy := make([]string, len(commonTpls))
			copy(commonTplsCopy, commonTpls)
			viewContent := append(commonTplsCopy, data)
			s.htmlRenderer.AddFromStringsFuncs(viewName, s.viewHelper, viewContent...)
		}
	}

	return nil
}

func getCommonTemplates(assets http.FileSystem, build, path string) ([]string, error) {
	var (
		fis []os.FileInfo
		err error
	)

	tpls := []string{}
	if build == "debug" {
		fis, _ = ioutil.ReadDir(path)
	} else {
		var file http.File
		path = "/" + path

		if file, err = assets.Open(path); err != nil {
			return nil, err
		}

		fis, _ = file.Readdir(-1)
	}

	for _, fi := range fis {
		if fi.IsDir() == true {
			continue
		}

		data, err := getTemplateContent(assets, build, path+"/"+fi.Name())
		if err != nil {
			return nil, err
		}

		tpls = append(tpls, data)
	}

	return tpls, nil
}

func getTemplateContent(assets http.FileSystem, build, path string) (string, error) {
	var data []byte
	if build == "debug" {
		data, _ := ioutil.ReadFile(path)
		return string(data), nil
	}

	file, err := assets.Open(path)
	if err != nil {
		return "", err
	}

	data, _ = ioutil.ReadAll(file)
	return string(data), nil
}

func isSSRPath(routes []RouteInfo, path string) bool {
	for _, route := range routes {
		if route.Path == path {
			return true
		}
	}

	return false
}
