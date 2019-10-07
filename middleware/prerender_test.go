package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type PrerenderSuiteT struct {
	test.SuiteT
	Config   *support.ConfigT
	Recorder *httptest.ResponseRecorder
}

func (s *PrerenderSuiteT) SetupTest() {
	s.Config = &support.ConfigT{}
	support.Copy(&s.Config, &support.Config)
	s.Recorder = httptest.NewRecorder()
}

func (s *PrerenderSuiteT) TestRequestWithNonSEOBot() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	output := support.CaptureOutput(func() {
		Prerender(s.Config)(ctx)
	})

	support.Logger.Info(output)
	s.NotContains(output, "SEO bot")
}

func (s *PrerenderSuiteT) TestRequestHTTPHostWithSEOBot() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "0.0.0.0",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	output := support.CaptureOutput(func() {
		Prerender(s.Config)(ctx)
	})

	support.Logger.Info(output)
	s.Contains(output, "SEO bot \"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)\" crawling \"http://0.0.0.0:3000/\"...")
	s.Equal(200, ctx.Writer.Status())
}

func (s *PrerenderSuiteT) TestRequestHTTPSHostWithSEOBot() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "0.0.0.0",
		Method: "GET",
		URL: &url.URL{
			Path: "/tools/about",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	s.Config.HTTPSSLEnabled = true
	output := support.CaptureOutput(func() {
		Prerender(s.Config)(ctx)
	})

	support.Logger.Info(output)
	s.Contains(output, "SEO bot \"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)\" crawling \"https://0.0.0.0:3443/tools/about\"...")
	s.Equal(200, ctx.Writer.Status())
}

func (s *PrerenderSuiteT) TestRequestFailedWithSEOBot() {
	oldCrawl := crawl
	crawl = func(url string) ([]byte, error) {
		return nil, errors.New("crawl failed")
	}
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "0.0.0.0",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	output := support.CaptureOutput(func() {
		Prerender(s.Config)(ctx)
	})

	support.Logger.Info(output)
	s.Contains(output, "SEO bot \"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)\" crawling \"http://0.0.0.0:3000/\"...")
	s.Equal(500, ctx.Writer.Status())
	crawl = oldCrawl
}

func TestPrerender(t *testing.T) {
	test.Run(t, new(PrerenderSuiteT))
}
