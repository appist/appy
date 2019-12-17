package http

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/appist/appy/support"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

var (
	_staticExtRegex = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
	userAgentHeader = http.CanonicalHeaderKey("user-agent")
	xPrerender      = http.CanonicalHeaderKey("x-prerender")
	crawl           = func(url string) ([]byte, error) {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.NoSandbox,
			chromedp.Headless,
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
)

// Prerender dynamically renders client-side rendered SPA for SEO using Chrome.
func Prerender(config *support.Config, logger *support.Logger) HandlerFunc {
	scheme := "http"
	host := config.HTTPHost
	port := config.HTTPPort

	if config.HTTPSSLEnabled == true {
		scheme = "https"
		port = config.HTTPSSLPort
	}

	return func(ctx *Context) {
		request := ctx.Request
		userAgent := request.Header.Get(userAgentHeader)

		if !_staticExtRegex.MatchString(request.URL.Path) && isSEOBot(userAgent) {
			url := fmt.Sprintf("%s://%s:%s%s", scheme, host, port, request.URL)
			logger.Infof("SEO bot \"%s\" crawling \"%s\"...", userAgent, url)

			data, err := crawl(url)
			if err != nil {
				logger.Error(err)
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			ctx.Writer.Header().Add(xPrerender, "1")
			ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
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
