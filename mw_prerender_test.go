package appy

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type PrerenderSuite struct {
	TestSuite
	config   *Config
	logger   *Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
}

func (s *PrerenderSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = newMockedLogger()
	s.config = NewConfig(DebugBuild, s.logger, nil)
	s.recorder = httptest.NewRecorder()
}

func (s *PrerenderSuite) TearDownTest() {
	os.Clearenv()
}

func (s *PrerenderSuite) TestRequestWithNonSEOBot() {
	ctx, _ := CreateTestContext(s.recorder)
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
	ctx, _ := CreateTestContext(s.recorder)
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
	ctx, _ := CreateTestContext(s.recorder)
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
	ctx, _ := CreateTestContext(s.recorder)
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
	RunTestSuite(t, new(PrerenderSuite))
}
