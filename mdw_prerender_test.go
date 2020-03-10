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
	asset    *Asset
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

	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata"), nil, "")
	s.config = NewConfig(s.asset, s.logger, &Support{})
	s.recorder = httptest.NewRecorder()
}

func (s *PrerenderSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *PrerenderSuite) TestRequestWithNonSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	c.Set(crawlerCtxKey.String(), &Crawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func (s *PrerenderSuite) TestRequestHTTPHostWithSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	c.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(c)
	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))

	c.Set(crawlerCtxKey.String(), &Crawl{})
	Prerender(s.config, s.logger)(c)
	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))
}

func (s *PrerenderSuite) TestRequestHTTPSHostWithSEOBot() {
	s.config.HTTPSSLEnabled = true
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/tools/about",
		},
	}
	c.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(c)
	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))

	c.Set(crawlerCtxKey.String(), &Crawl{})
	Prerender(s.config, s.logger)(c)
	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))
}

type mockCrawl struct{}

func (m mockCrawl) Perform(url string) ([]byte, error) {
	return nil, errors.New("crawl failed")
}

func (s *PrerenderSuite) TestRequestFailedWithSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Host:   "localhost",
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	c.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	c.Set(crawlerCtxKey.String(), &mockCrawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(500, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func TestPrerenderSuite(t *testing.T) {
	RunTestSuite(t, new(PrerenderSuite))
}
