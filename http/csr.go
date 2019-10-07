package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

var (
	// CSRRoot is the root folder for VueJS web application.
	CSRRoot         = "web"
	extRegex        = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
	userAgentHeader = http.CanonicalHeaderKey("user-agent")
)

// InitCSR setup the SPA client-side rendering/routing with index.html fallback.
func (s *ServerT) InitCSR() {
	// Setup SPA hosting at "/".
	s.Router.Use(serveSPA("/", s))

	s.Router.NoRoute(middleware.CSRFSkipCheck(), func(c *gin.Context) {
		request := c.Request
		method := request.Method
		path := request.URL.Path
		fallback := !extRegex.MatchString(path)

		// Prevent SEO crawling from double rendering.
		if isSEOBot(request.Header.Get(userAgentHeader)) {
			return
		}

		if method == "GET" && fallback {
			c.Header("Cache-Control", "no-cache")

			data := []byte{}
			if s.Assets != nil {
				f, _ := s.Assets.Open("/index.html")
				if f != nil {
					data, _ = ioutil.ReadAll(f)
				}
			}

			if len(data) > 0 {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}
		}

		c.HTML(http.StatusNotFound, "error/404", gin.H{
			"title": "404 Page Not Found",
		})
	})
}

func serveSPA(url string, s *ServerT) HandlerFuncT {
	fs := http.FileServer(s.Assets)
	if url != "" {
		fs = http.StripPrefix(url, fs)
	}

	return func(c *ContextT) {
		request := c.Request

		if !support.Contains([]string{"/" + SSRView + "/", "/" + SSRLocale + "/"}, request.URL.Path) &&
			!isSSRPath(s.Routes(), request.URL.Path) {
			userAgent := request.Header.Get(userAgentHeader)

			// SEO crawling for the CSR pages.
			if !extRegex.MatchString(request.URL.Path) && isSEOBot(userAgent) {
				scheme := "http"
				port := s.Config.HTTPPort
				if s.Config.HTTPSSLEnabled == true {
					scheme = "https"
					port = s.Config.HTTPSSLPort
				}

				targetURL := fmt.Sprintf("%s://%s:%s%s", scheme, s.Config.HTTPHost, port, request.URL)
				support.Logger.Infof("SEO bot \"%s\" crawling \"%s\"...", userAgent, targetURL)

				data, err := crawl(targetURL)
				if err != nil {
					support.Logger.Error(err)
					c.AbortWithError(http.StatusInternalServerError, err)
					return
				}

				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}

			if s.Assets != nil {
				_, err := s.Assets.Open(request.URL.Path)
				if err == nil {
					fs.ServeHTTP(c.Writer, request)
					c.Abort()
				}
			}
		}
	}
}

func crawl(url string) ([]byte, error) {
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
		chromedp.Navigate(url),
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

	return []byte(data), err
}

func isSSRPath(routes []RouteInfoT, path string) bool {
	for _, route := range routes {
		if route.Path == path {
			return true
		}
	}

	return false
}

func isSEOBot(ua string) bool {
	bots := []string{
		"googlebot", "yahoou", "bingbot", "baiduspider", "yandex", "yeti", "yodaobot", "gigabot", "ia_archiver",
		"facebookexternalhit", "twitterbot", "developers.google.com",
	}

	for _, bot := range bots {
		if strings.Contains(strings.ToLower(ua), bot) {
			return true
		}
	}

	return false
}
