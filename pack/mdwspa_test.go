package pack

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwSPASuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	logger *support.Logger
	server *Server
}

func (s *mdwSPASuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "testdata/mdwspa")
	s.config = support.NewConfig(s.asset, s.logger)
	s.server = NewServer(s.asset, s.config, s.logger)
	s.server.GET("/ssr", func(c *Context) {
		c.Data(http.StatusOK, "text/html", []byte("server-side rendering"))
	})
}

func (s *mdwSPASuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwSPASuite) TestSSROrReservedPath() {
	spa := mdwSPA(s.server, "/spa", http.Dir("testdata/mdwspa"))
	urls := []string{
		"/ssr",
		"/" + s.asset.Layout().Config(),
		"/docker-compose.yml",
		"/" + s.asset.Layout().Locale(),
		"/" + s.asset.Layout().View(),
	}
	for _, u := range urls {
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: u}}
		spa(c)

		s.Equal(false, c.IsAborted())
	}
}

func (s *mdwSPASuite) TestDebugBuild() {
	{
		spa := mdwSPA(s.server, "/spa", http.Dir("testdata/mdwspa"))
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/spa"}}
		spa(c)

		s.Equal(true, c.IsAborted())
	}

	{
		s.server.config.HTTPSSLEnabled = true
		spa := mdwSPA(s.server, "/spa", http.Dir("testdata/mdwspa"))
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/spa"}}
		spa(c)

		s.Equal(true, c.IsAborted())
	}
}

func (s *mdwSPASuite) TestReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() { support.Build = support.DebugBuild }()

	{
		spa := mdwSPA(s.server, "/spa", nil)
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/app.js"}}
		spa(c)

		s.Equal(false, c.IsAborted())
	}

	{
		spa := mdwSPA(s.server, "/spa", nil)
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/spa"}}
		spa(c)

		s.Equal(false, c.IsAborted())
	}

	{
		spa := mdwSPA(s.server, "/spa", http.Dir("testdata/mdwspa/missing"))
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/spa"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal(http.StatusInternalServerError, recorder.Code)
	}

	{

		spa := mdwSPA(s.server, "/", &fakeFS{})
		recorder := httptest.NewRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/login"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal("fake error", c.Errors.Last().Error())
		s.Equal(http.StatusInternalServerError, recorder.Code)
	}

	{
		spa := mdwSPA(s.server, "/", http.Dir("testdata/mdwspa"))
		recorder := httptest.NewRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/login"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "<title>Single Page Application</title>")
	}

	{
		spa := mdwSPA(s.server, "/", http.Dir("testdata/mdwspa"))
		recorder := httptest.NewRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/app.js"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "func app() {}")
	}

	{
		spa := mdwSPA(s.server, "/admin", http.Dir("testdata/mdwspa/admin"))
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/admin/login"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "<title>Single Page Application - Admin</title>")
	}

	{
		spa := mdwSPA(s.server, "/admin", http.Dir("testdata/mdwspa/admin"))
		recorder := NewResponseRecorder()
		c, _ := NewTestContext(recorder)
		c.Request = &http.Request{URL: &url.URL{Path: "/admin/app.js"}}
		spa(c)

		s.Equal(true, c.IsAborted())
		s.Equal(http.StatusOK, recorder.Code)
		s.Contains(recorder.Body.String(), "func adminApp() {}")
	}
}

func TestMdwSPASuite(t *testing.T) {
	test.Run(t, new(mdwSPASuite))
}

type fakeFile struct{}

func (f *fakeFile) Close() error {
	return nil
}

func (f *fakeFile) Read(p []byte) (n int, err error) {
	return 0, errors.New("fake error")
}

func (f *fakeFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f fakeFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *fakeFile) Stat() (os.FileInfo, error) {
	return nil, nil
}

type fakeFS struct{}

func (f *fakeFS) Open(name string) (http.File, error) {
	return &fakeFile{}, nil
}
