package appy

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
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

type (
	// H is a type alias to gin.H.
	H = gin.H

	// HandlerFunc is a type alias to gin.HandlerFunc.
	HandlerFunc = gin.HandlerFunc

	// Context is a type alias to gin.Context.
	Context = gin.Context

	// HandlersChain is a type alias to gin.HandlersChain.
	HandlersChain = gin.HandlersChain

	// Router is a type alias to gin.Engine.
	Router = gin.Engine

	// RouterGroup is a type alias to gin.RouterGroup.
	RouterGroup = gin.RouterGroup

	// RouteInfo is a type alias to gin.RouteInfo.
	RouteInfo = gin.RouteInfo

	// Routes is a type alias to gin.IRoutes.
	Routes = gin.IRoutes

	csrResource struct {
		assets     http.FileSystem
		fileServer http.Handler
		prefix     string
	}

	// Server is the engine that serves HTTP/GRPC requests.
	Server struct {
		assets       http.FileSystem
		config       *Config
		csrResources []csrResource
		grpc         *grpc.Server
		http         *http.Server
		htmlRenderer multitemplate.Renderer
		i18nBundle   *i18n.Bundle
		logger       *Logger
		router       *Router
		csrPaths     map[string]string
		ssrPaths     map[string]string
		viewHelper   template.FuncMap
	}
)

var (
	reservedViewDirs = []string{"layouts", "shared"}
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewServer initializes Server instance.
func NewServer(c *Config, l *Logger, assets http.FileSystem, viewHelper template.FuncMap) *Server {
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

	return &Server{
		assets:       assets,
		config:       c,
		csrResources: []csrResource{},
		grpc:         nil,
		http:         http,
		htmlRenderer: renderer,
		logger:       l,
		router:       r,
		csrPaths:     _csrPaths,
		ssrPaths:     _ssrPaths,
		viewHelper:   viewHelper,
	}
}

// Hosts returns the server hosts list.
func (s Server) Hosts() ([]string, error) {
	var hosts = []string{s.config.HTTPHost}

	if s.config.HTTPHost != "localhost" {
		hosts = append(hosts, "localhost")
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

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
func (s Server) IsSSLCertExisted() bool {
	_, certErr := os.Stat(s.config.HTTPSSLCertPath + "/cert.pem")
	_, keyErr := os.Stat(s.config.HTTPSSLCertPath + "/key.pem")

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return false
	}

	return true
}

// PrintInfo prints the server info.
func (s Server) PrintInfo() {
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			VERSION, runtime.Version(), Build, s.config.AppyEnv, s.config.path,
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.config.HTTPPort)

	if s.config.HTTPSSLEnabled == true {
		host = fmt.Sprintf("https://%s:%s", hosts[0], s.config.HTTPSSLPort)
	}

	lines = append(lines, fmt.Sprintf("* Listening on %s", host))

	for _, line := range lines {
		s.logger.Info(line)
	}
}

// Routes returns all the routes including those in middlewares.
func (s Server) Routes() []RouteInfo {
	routes := s.router.Routes()

	if s.config.HTTPHealthCheckURL != "" {
		routes = append(routes, RouteInfo{
			Method:      "GET",
			Path:        s.config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// SetRoutes configures routes for the server.
func (s Server) SetRoutes(cb func(router *Router)) {
	cb(s.router)
}

// InitCSR initializes the client-side rendering/routing with index.html fallback.
func (s Server) InitCSR() {
	// Setup CSR hosting at "/".
	s.router.Use(s.serveCSR("/", s.assets))

	// Setup CSR hosting at "/tools".
	// s.Router.Use(s.serveCSR("/tools", tools.Assets))

	// Note: This might be not needed once we're sure everything is supposed to be on the PWA.
	s.router.NoRoute(CSRFSkipCheck(), func(ctx *Context) {
		ctx.HTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

func (s Server) csrResource(path string) csrResource {
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

func (s *Server) serveCSR(prefix string, assets http.FileSystem) HandlerFunc {
	s.csrResources = append(s.csrResources, csrResource{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(ctx *Context) {
		request := ctx.Request

		// Serve from the webpack-dev-server(debug) or assets(release) if the URL path isn't matching any of the SSR
		// paths.
		if !isSSRPath(s.Routes(), request.URL.Path) && !strings.HasPrefix(request.URL.Path, "/"+s.ssrPaths["root"]) {
			resource := s.csrResource(request.URL.Path)

			if Build == DebugBuild {
				director := func(req *http.Request) {
					port, _ := strconv.Atoi(s.config.HTTPPort)
					req.URL.Scheme = "http"
					if s.config.HTTPSSLEnabled {
						port, _ = strconv.Atoi(s.config.HTTPSSLPort)
						req.URL.Scheme = "https"
					}

					hostname := s.config.HTTPHost + ":" + strconv.Itoa(port+1)
					req.URL.Host = hostname
					req.Host = hostname
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(ctx.Writer, ctx.Request)
				ctx.Abort()
				return
			}

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
func (s Server) InitSSR() error {
	if err := s.initSSRLocale(); err != nil {
		return err
	}

	if err := s.initSSRView(); err != nil {
		return err
	}

	return nil
}

func (s *Server) initSSRLocale() error {
	s.i18nBundle = i18n.NewBundle(language.English)
	s.i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var (
		localeFiles []os.FileInfo
		data        []byte
		err         error
	)
	localeDir := s.ssrPaths["locale"]

	// Try getting all the locale files from `pkg/locales`, but fallback to `assets` http.FileSystem.
	if Build == "debug" {
		localeFiles, err = ioutil.ReadDir(localeDir)

		if err != nil {
			return err
		}
	} else {
		localeDir = "/" + s.ssrPaths["root"] + "/" + localeDir
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

	s.router.Use(I18n(s.i18nBundle))
	return nil
}

func (s *Server) initSSRView() error {
	var (
		fis []os.FileInfo
		err error
	)

	viewDir := s.ssrPaths["view"]

	// We will always read from local file system when it's debug build. Otherwise, read from the bind assets.
	if Build == "debug" {
		if fis, err = ioutil.ReadDir(viewDir); err != nil {
			return err
		}
	} else {
		viewDir = "/" + s.ssrPaths["root"] + "/" + viewDir

		var file http.File
		if file, err = s.assets.Open(viewDir); err != nil {
			return err
		}

		fis, _ = file.Readdir(-1)
	}

	commonTpls := []string{}
	for _, fi := range fis {
		// We should only see directories in `pkg/views`.
		if fi.IsDir() == false {
			continue
		}

		if ArrayContains(reservedViewDirs, fi.Name()) == true {
			tpls, err := getCommonTemplates(s.assets, Build, viewDir+"/"+fi.Name())
			if err != nil {
				return err
			}

			commonTpls = append(commonTpls, tpls...)
		}
	}

	for _, fi := range fis {
		if fi.IsDir() == false || ArrayContains(reservedViewDirs, fi.Name()) == true {
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

func newRouter(c *Config, l *Logger) *Router {
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

func newSecureConfig(c *Config) secure.Config {
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

func errorTplUpper() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<title>{{.title}}</title>
			<link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" />
			<style>
			body { padding-top: 4.5rem; }
			</style>
	  	</head>
		<body>
			<nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark">
				<div class="navbar-brand">{{.title}}</div>
			</nav>
			<main role="main" class="px-3">
	`
}

func errorTplLower() string {
	return `
			</main>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
			<script src="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/js/bootstrap.min.js"></script>
		</body>
	</html>
	`
}

func errorTpl404() string {
	return errorTplUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">The page that you are looking for does not exist, please contact the website administrator for more details.</p>
	</div>
</div>
		` + errorTplLower()
}

func errorTpl500() string {
	if Build == "debug" {
		return errorTplUpper() + `
<h2 class="text-danger">Full Trace</h2>
<pre class="pre-scrollable bg-light p-2">{{range $error := .errors}}{{$error}}{{end}}</pre>
<h2 class="text-danger">Request</h2>
<h6>Headers</h6>
<pre class="pre-scrollable bg-light p-2">{{.headers}}</pre>
<h6>Query String Parameters</h6>
<pre class="pre-scrollable bg-light p-2">{{.qsParams}}</pre>
<h6>Session Variables</h6>
<pre class="pre-scrollable bg-light p-2">{{.sessionVars}}</pre>
		` + errorTplLower()
	}

	return errorTplUpper() + `
<div class="card mx-auto bg-light" style="max-width:30rem;margin-top:3rem;">
	<div class="card-body">
		<p class="card-text">If you are the administrator of this website, then please read this web application's log file and/or the web server's log file to find out what went wrong.</p>
	</div>
</div>
	` + errorTplLower()
}

func welcomeTpl() string {
	return `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
			<title>{{.title}}</title>
			<link href="//cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" />
			<style>
				/*
				* Base structure
				*/
				html,
				body {
					height: 100%;
				}
				body {
					display: -ms-flexbox;
					display: flex;
				}
				.cover-container {
					max-width: 48em;
				}
				/*
				* Header
				*/
				.masthead {
					margin-bottom: 2rem;
				}
				.masthead-brand {
					margin-bottom: 0;
				}
				@media (min-width: 48em) {
					.masthead-brand {
						float: left;
					}
				}
				/*
				* Cover
				*/
				.cover {
					margin-top: -3rem;
					padding: 0 1.5rem;
				}
				.cover .lead:first-of-type {
					margin-top: -3rem;
					margin-bottom: 1.5rem;
				}
				/*
				* Footer
				*/
				.mastfoot {
					color: rgba(255, 255, 255, .5);
				}
			</style>
		</head>
		<body class="text-center">
			<div class="cover-container d-flex w-100 h-100 p-2 mx-auto flex-column">
				<header class="masthead mb-auto"></header>
				<main role="main" class="inner cover">
					<h1 class="cover-heading">` + logoImage() + `</h1>
					<p class="lead">An opinionated productive web framework that helps scaling business easier.</p>
					<p class="lead">
						<a href="https://appy.appist.io" class="btn btn-lg btn-primary">Learn more</a>
					</p>
				</main>
				<footer class="mastfoot mt-auto"></footer>
			</div>
		</body>
	</html>
	`
}

func logoImage() string {
	return `
	<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="18rem" height="18rem" viewBox="0 0 1024 1024" xml:space="preserve">
		<g transform="matrix(1 0 0 1 512 512)" id="background-logo">
			<rect style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(255,255,255); fill-rule: nonzero; opacity: 1;"  paint-order="stroke"  x="-512" y="-512" rx="0" ry="0" width="1024" height="1024" />
		</g>
		<g transform="matrix(0.45660522273425497 0 0 -0.45660522273425497 512 395.2665130568356)" id="maker-logo">
			<path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(44,123,229); fill-rule: nonzero; opacity: 1;"  paint-order="stroke"  transform=" translate(-815.895, -767.6175000000001)" d="m 1082.67 409.402 l -55.05 -31.785 V 708.918 L 815.895 587.73 L 604.176 708.918 V 377.617 l -55.059 31.785 v 594.508 l 55.059 -31.785 V 757.188 L 815.895 634.949 L 1027.62 757.188 v 215.21 l 55.05 31.782 z M 841.746 864.5 L 975.906 941.957 V 787.043 l -134.16 -77.457 z m -51.707 0 V 709.586 l -134.156 77.457 v 154.914 z m -108.301 122.234 l 134.157 77.456 l 134.16 -77.456 l -134.16 -77.453 z m 134.157 200.736 l 240.415 -138.8 l -55.06 -31.78 l -185.355 107.01 l -185.594 -107.15 l -55.055 31.79 z m 0 59.71 L 497.41 1063.3 V 379.555 l 106.766 -61.645 l 51.707 -29.855 v 59.707 v 272.515 l 160.012 -92.386 l 160.011 92.386 V 347.766 V 288.059 l 51.714 29.855 l 106.76 61.641 V 1063.3 l -318.485 183.88" stroke-linecap="round" />
		</g>
		<g transform="matrix(1.5745007680491552 0 0 1.5745007680491552 512.0074881042386 703.06159621802)">
			<filter id="SVGID_14400" y="-20%" height="140%" x="-20%" width="140%">
				<feGaussianBlur in="SourceAlpha" stdDeviation="0"></feGaussianBlur>
				<feOffset dx="0" dy="0" result="oBlur" ></feOffset>
				<feFlood flood-color="rgb(0,0,0)" flood-opacity="1"/>
				<feComposite in2="oBlur" operator="in" />
				<feMerge>
					<feMergeNode></feMergeNode>
					<feMergeNode in="SourceGraphic"></feMergeNode>
				</feMerge>
			</filter>
			<path style="stroke: none; stroke-width: 1; stroke-dasharray: none; stroke-linecap: butt; stroke-dashoffset: 0; stroke-linejoin: miter; stroke-miterlimit: 4; fill: rgb(44,123,229); fill-rule: nonzero; opacity: 1;filter: url(#SVGID_14400);"  paint-order="stroke"  transform=" translate(-101.43999999999998, 12.774999999999999)" d="M 29.05 0 L 6.24 0 L 6.24 -22.71 L 34.63 -22.71 L 34.63 -31.24 L 17.61 -31.24 L 17.61 -28.39 L 6.24 -28.39 L 6.24 -34.43 L 6.24 -39.76 L 40.18 -39.76 Q 41.19 -38.83 45.94 -34.08 L 45.94 -34.08 L 45.94 0 L 34.63 0 L 34.63 -2.84 Q 33.21 -2.11 29.05 0 L 29.05 0 Z M 17.61 -11.72 L 17.61 -8.53 L 28.98 -8.53 Q 29.54 -8.81 34.63 -11.3 L 34.63 -11.3 L 34.63 -14.21 L 26.14 -14.21 L 17.61 -14.21 L 17.61 -11.72 Z M 58.48 14.21 L 58.48 -13.83 L 58.48 -39.76 L 64.17 -39.76 L 69.86 -39.76 L 69.86 -34.08 L 78.38 -39.76 L 98.18 -39.76 L 98.18 -18.44 L 98.18 -11.37 L 98.18 -5.69 L 96.13 -3.54 L 92.56 0 L 69.86 0 L 69.86 14.21 L 58.48 14.21 Z M 69.86 -17.06 L 69.86 -8.53 L 86.88 -8.53 L 86.88 -31.24 L 78.38 -31.24 Q 76.58 -29.95 74.05 -28.39 L 74.05 -28.39 Q 73.7 -28.19 72.2 -27.15 Q 70.69 -26.1 69.86 -25.55 L 69.86 -25.55 L 69.86 -17.06 Z M 110.73 14.21 L 110.73 -13.83 L 110.73 -39.76 L 116.42 -39.76 L 122.1 -39.76 L 122.1 -34.08 L 130.63 -39.76 L 150.42 -39.76 L 150.42 -18.44 L 150.42 -11.37 L 150.42 -5.69 L 148.38 -3.54 L 144.81 0 L 122.1 0 L 122.1 14.21 L 110.73 14.21 Z M 122.1 -17.06 L 122.1 -8.53 L 139.12 -8.53 L 139.12 -31.24 L 130.63 -31.24 Q 128.83 -29.95 126.3 -28.39 L 126.3 -28.39 Q 125.95 -28.19 124.44 -27.15 Q 122.93 -26.1 122.1 -25.55 L 122.1 -25.55 L 122.1 -17.06 Z M 185.27 -11.37 L 162.56 -11.37 L 162.56 -25.55 L 162.56 -39.76 L 173.93 -39.76 L 173.93 -34.08 L 173.93 -22.71 L 185.27 -22.71 L 185.27 -39.76 L 196.64 -39.76 L 196.64 8.53 L 168.24 8.53 L 168.24 0 L 185.27 0 L 185.27 -11.37 Z" stroke-linecap="round" />
		</g>
	</svg>
`
}
