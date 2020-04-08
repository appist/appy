package pack

import (
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/appist/appy/support"
)

func mdwSPA(server *Server, prefix string, fs http.FileSystem) HandlerFunc {
	server.spaResources = append(server.spaResources, &spaResource{
		fs:         fs,
		fileServer: http.StripPrefix(prefix, http.FileServer(fs)),
		prefix:     prefix,
	})

	return func(c *Context) {
		req := c.Request

		if server.isSSRPath(req.URL.Path) ||
			strings.HasPrefix(req.URL.Path, "/"+server.asset.Layout().Config()) ||
			strings.HasPrefix(req.URL.Path, "/"+server.asset.Layout().Docker()) ||
			strings.HasPrefix(req.URL.Path, "/"+server.asset.Layout().Locale()) ||
			strings.HasPrefix(req.URL.Path, "/"+server.asset.Layout().View()) {
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
		if resource == nil || resource.fs == nil {
			c.Next()
			return
		}

		// If the request path is a pretty URL, serve the index.html from the prefix as fallback.
		if filepath.Ext(req.URL.Path) == "" {
			file, err := resource.fs.Open("/index.html")
			if err != nil {
				server.logger.Error(err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			data, err := ioutil.ReadAll(file)
			if err != nil {
				server.logger.Error(err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			c.Abort()
			return
		}

		// Only serve the request from fs if the file is in the assets filesystem.
		resource.fileServer.ServeHTTP(c.Writer, req)
		c.Abort()
	}
}
