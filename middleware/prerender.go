package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/appist/appy/support"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

var (
	userAgentHeader = http.CanonicalHeaderKey("user-agent")
)

// Prerender dynamically renders client-side rendered SPA for SEO using Chrome.
func Prerender(config *support.ConfigT) gin.HandlerFunc {
	scheme := "http"
	host := config.HTTPHost
	port := config.HTTPPort
	if config.HTTPSSLEnabled == true {
		scheme = "https"
		port = config.HTTPSSLPort
	}

	return func(c *gin.Context) {
		request := c.Request
		userAgent := request.Header.Get(userAgentHeader)

		if !support.ExtRegex.MatchString(request.URL.Path) && isSEOBot(userAgent) {
			url := fmt.Sprintf("%s://%s:%s%s", scheme, host, port, request.URL)
			support.Logger.Infof("SEO bot \"%s\" crawling \"%s\"...", userAgent, url)

			data, err := crawl(url)
			if err != nil {
				support.Logger.Error(err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
			return
		}

		c.Next()
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
