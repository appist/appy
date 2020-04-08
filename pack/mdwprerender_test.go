package pack

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwPrerenderSuite struct {
	test.Suite
	asset    *support.Asset
	config   *support.Config
	logger   *support.Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
}

func (s *mdwPrerenderSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.recorder = httptest.NewRecorder()
}

func (s *mdwPrerenderSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwPrerenderSuite) TestRequestWithNonSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/",
		},
	}
	c.Set(mdwPrerenderCtxKey.String(), &crawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func (s *mdwPrerenderSuite) TestNonGETRequestWithSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		URL: &url.URL{
			Path: "/",
		},
	}
	c.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func (s *mdwPrerenderSuite) TestRequestStaticExtWithSEOBot() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
		URL: &url.URL{
			Path: "/app.js",
		},
	}
	c.Request.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func (s *mdwPrerenderSuite) TestRequestHTTPHostWithSEOBot() {
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

	c.Set(mdwPrerenderCtxKey.String(), &crawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))
}

func (s *mdwPrerenderSuite) TestRequestHTTPSHostWithSEOBot() {
	s.config.HTTPSSLEnabled = true
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

	c.Set(mdwPrerenderCtxKey.String(), &crawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(http.StatusOK, c.Writer.Status())
	s.Equal("1", c.Writer.Header().Get(xPrerender))
}

func (s *mdwPrerenderSuite) TestRequestFailedWithSEOBot() {
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
	c.Set(mdwPrerenderCtxKey.String(), &mockCrawl{})
	Prerender(s.config, s.logger)(c)

	s.Equal(500, c.Writer.Status())
	s.Equal("", c.Writer.Header().Get(xPrerender))
}

func TestMdwPrerenderSuite(t *testing.T) {
	test.Run(t, new(mdwPrerenderSuite))
}

type mockCrawl struct{}

func (m mockCrawl) Perform(url string) ([]byte, error) {
	return nil, errors.New("crawl failed")
}
