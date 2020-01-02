package http

import (
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/appist/appy/support"
)

// SPA is the middleware that sets up the SPA hosting at the specified prefix path.
func SPA(server *Server, prefix string, assets http.FileSystem) HandlerFunc {
	server.spaResources = append(server.spaResources, &spaResource{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(c *Context) {
		req := c.Request
		if server.isSSRPath(req.URL.Path) || strings.HasPrefix(req.URL.Path, "/"+server.assets.SSRRelease()) || strings.HasPrefix(req.URL.Path, "/"+server.assets.Layout()["config"]) {
			c.Next()
			return
		}

		// Serve from the webpack-dev-server for debug build.
		if support.IsDebugBuild() {
			director := func(req *http.Request) {
				port, _ := strconv.Atoi(server.config.HTTPPort)
				req.URL.Scheme = "http"
				if server.config.HTTPSSLEnabled {
					port, _ = strconv.Atoi(server.config.HTTPSSLPort)
					req.URL.Scheme = "https"
				}

				hostname := server.config.HTTPHost + ":" + strconv.Itoa(port+1)
				req.URL.Host = hostname
				req.Host = hostname
			}
			proxy := &httputil.ReverseProxy{Director: director}
			proxy.ServeHTTP(c.Writer, req)
			c.Abort()
			return
		}

		resource := server.spaResource(req.URL.Path)
		if resource == nil || resource.assets == nil {
			c.Next()
			return
		}

		// Only serve the request from assets if the file is in the assets filesystem.
		resource.fileServer.ServeHTTP(c.Writer, req)
		c.Abort()
	}
}
