package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"appist/appy/middleware"
	"appist/appy/support"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

var (
	extRegex        = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
	headerUserAgent = http.CanonicalHeaderKey("user-agent")
)

// SetupCSR sets up the server for client-side rendering with prerender enabled.
func (s *ServerT) SetupCSR() {
	s.router.Use(serveStaticAssets("/", s))
	s.router.NoRoute(middleware.CSRFSkipCheck(), func(c *gin.Context) {
		var data []byte
		r := c.Request
		method := r.Method
		path := r.URL.Path
		fallback := !extRegex.MatchString(path)

		if isWebCrawler(strings.ToLower(r.Header.Get(headerUserAgent))) {
			return
		}

		if method == "GET" && fallback && s.assets != nil {
			c.Header("Cache-Control", "no-cache")

			f, err := s.assets.Open("/index.html")
			if err != nil {
				c.HTML(http.StatusNotFound, "error/404", gin.H{
					"title": "404 Page Not Found",
				})
				return
			}

			data, _ = ioutil.ReadAll(f)
			s.RenderData(c, http.StatusOK, "text/html; charset=utf-8", data)
			return
		}

		// TODO: Add different Content-Type handling.
		c.HTML(http.StatusNotFound, "error/404", gin.H{
			"title": "404 Page Not Found",
		})
	})
}

func serveStaticAssets(urlPrefix string, s *ServerT) gin.HandlerFunc {
	fileserver := http.FileServer(s.assets)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}

	routesMap := map[string]bool{}
	for _, route := range s.router.Routes() {
		if route.Method == "GET" {
			routesMap[route.Path] = true
		}
	}

	return func(c *gin.Context) {
		r := c.Request
		if !strings.Contains(r.URL.Path, "/"+ViewDir+"/") &&
			!strings.Contains(r.URL.Path, "/"+LocaleDir+"/") &&
			!isSSRPath(routesMap, r.URL.Path) {
			// SEO crawling for the CSR pages.
			if !extRegex.MatchString(r.URL.Path) &&
				isWebCrawler(strings.ToLower(r.Header.Get(headerUserAgent))) {
				scheme := "http"
				port := support.Config.HTTPPort
				if support.Config.HTTPSSLEnabled == true {
					scheme = "https"
					port = support.Config.HTTPSSLPort
				}

				urlToCrawl := fmt.Sprintf("%s://%s:%s%s", scheme, support.Config.HTTPHost, port, r.RequestURI)
				support.Logger.Infof("SEO crawling %s...", urlToCrawl)

				opts := append(chromedp.DefaultExecAllocatorOptions[:],
					chromedp.DisableGPU,
					chromedp.NoSandbox,
					chromedp.Flag("ignore-certificate-errors", true),
				)

				allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
				defer cancel()

				taskCtx, cancel := chromedp.NewContext(allocCtx)
				defer cancel()

				var data string
				err := chromedp.Run(taskCtx,
					chromedp.Navigate(urlToCrawl),
					chromedp.WaitReady("body", chromedp.ByQuery),
					chromedp.ActionFunc(func(ctx context.Context) error {
						node, err := dom.GetDocument().Do(ctx)
						if err != nil {
							return err
						}

						data, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
						return err
					}),
				)

				if err != nil {
					support.Logger.Error(err)
					c.AbortWithError(http.StatusInternalServerError, err)
				}

				s.RenderData(c, http.StatusOK, "text/html; charset=utf-8", []byte(data))
				return
			}

			if s.assets == nil {
				return
			}

			_, err := s.assets.Open(r.URL.Path)
			if err == nil {
				fileserver.ServeHTTP(c.Writer, r)
				c.Abort()
			}
		}
	}
}

func isSSRPath(routes map[string]bool, url string) bool {
	if _, ok := routes[url]; ok {
		return true
	}

	return false
}

func isWebCrawler(userAgent string) bool {
	normalisedUA := strings.ToLower(userAgent)
	if strings.Contains(normalisedUA, "googlebot") || strings.Contains(normalisedUA, "yahoou") || strings.Contains(normalisedUA, "bingbot") ||
		strings.Contains(normalisedUA, "baiduspider") || strings.Contains(normalisedUA, "yandex") || strings.Contains(normalisedUA, "yeti") ||
		strings.Contains(normalisedUA, "yodaobot") || strings.Contains(normalisedUA, "gigabot") || strings.Contains(normalisedUA, "ia_archiver") ||
		strings.Contains(normalisedUA, "facebookexternalhit") || strings.Contains(normalisedUA, "twitterbot") || strings.Contains(normalisedUA, "developers.google.com") {
		return true
	}

	return false
}
