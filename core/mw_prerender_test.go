package core

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
)

type PrerenderSuite struct {
	test.Suite
	config   AppConfig
	logger   *AppLogger
	recorder *httptest.ResponseRecorder
}

func (s *PrerenderSuite) SetupTest() {
	s.config, _ = newConfig(nil)
	s.config.HTTPSessionSecrets = [][]byte{[]byte("481e5d98a31585148b8b1dfb6a3c0465")}
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.logger, _ = newLogger(newLoggerConfig())
	s.recorder = httptest.NewRecorder()
}

func (s *PrerenderSuite) TearDownTest() {
}

func (s *PrerenderSuite) TestRequestWithNonSEOBot() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	Prerender(s.config, s.logger)(ctx)

	s.Equal(200, ctx.Writer.Status())
	s.Equal("", ctx.Writer.Header().Get(xPrerender))
}

func (s *PrerenderSuite) TestRequestHTTPHostWithSEOBot() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(ctx)

	s.Equal(200, ctx.Writer.Status())
	s.Equal("1", ctx.Writer.Header().Get(xPrerender))
}

func (s *PrerenderSuite) TestRequestHTTPSHostWithSEOBot() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/tools/about",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	s.config.HTTPSSLEnabled = true
	Prerender(s.config, s.logger)(ctx)

	s.Equal(200, ctx.Writer.Status())
	s.Equal("1", ctx.Writer.Header().Get(xPrerender))
}

func (s *PrerenderSuite) TestRequestFailedWithSEOBot() {
	oldCrawl := crawl
	crawl = func(url string) ([]byte, error) {
		return nil, errors.New("crawl failed")
	}
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	ctx.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(ctx)

	s.Equal(500, ctx.Writer.Status())
	s.Equal("", ctx.Writer.Header().Get(xPrerender))
	crawl = oldCrawl
}

func TestPrerender(t *testing.T) {
	test.Run(t, new(PrerenderSuite))
}
