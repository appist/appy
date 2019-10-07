package http

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/appist/appy/tools"
)

type spaResourceT struct {
	assets     http.FileSystem
	fileServer http.Handler
	prefix     string
}

var (
	// CSRRoot is the root folder for VueJS web application.
	CSRRoot         = "web"
	extRegex        = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
	userAgentHeader = http.CanonicalHeaderKey("user-agent")
	spaResources    = []spaResourceT{}
)

// InitCSR setup the SPA client-side rendering/routing with index.html fallback.
func (s *ServerT) InitCSR() {
	// Setup SPA hosting at "/".
	s.Router.Use(serveSPA("/", s.Assets, s))

	// Setup SPA hosting at "/tools".
	s.Router.Use(serveSPA("/tools", tools.Assets, s))

	s.Router.NoRoute(middleware.CSRFSkipCheck(), func(c *ContextT) {
		request := c.Request

		if request.Method == "GET" && !extRegex.MatchString(request.URL.Path) {
			c.Header("Cache-Control", "no-cache")
			resource := getSPAResource(request.URL.Path)

			if resource.assets != nil {
				file, _ := resource.assets.Open("/index.html")

				if file != nil {
					data, _ := ioutil.ReadAll(file)

					if data != nil {
						c.Data(http.StatusOK, "text/html; charset=utf-8", data)
						return
					}
				}
			}
		}

		c.HTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

func serveSPA(prefix string, assets http.FileSystem, s *ServerT) HandlerFuncT {
	spaResources = append(spaResources, spaResourceT{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(c *ContextT) {
		request := c.Request
		ssrAssetsPath := []string{"/" + SSRView + "/", "/" + SSRLocale + "/"}

		// Serve from the assets FS if the URL path isn't `/views/`, `/locales/` or any of the SSR path.
		if !support.Contains(ssrAssetsPath, request.URL.Path) && !isSSRPath(s.Routes(), request.URL.Path) {
			resource := getSPAResource(request.URL.Path)

			if resource.assets != nil {
				requestPath := request.URL.Path
				assetPath := strings.Replace(requestPath, resource.prefix, "", 1)
				_, err := resource.assets.Open(assetPath)
				// Only serve the request from assets if the file is in the assets filesystem.
				if err == nil {
					resource.fileServer.ServeHTTP(c.Writer, request)
					c.Abort()
				}
			}
		}
	}
}

func getSPAResource(path string) spaResourceT {
	var resource spaResourceT

	for _, spaResource := range spaResources {
		if (spaResourceT{}) == resource {
			resource = spaResource
			continue
		}

		if spaResource.prefix != "/" && strings.HasPrefix(path, spaResource.prefix) {
			resource = spaResource
			break
		}
	}

	return resource
}

func isSSRPath(routes []RouteInfoT, path string) bool {
	for _, route := range routes {
		if route.Path == path {
			return true
		}
	}

	return false
}
