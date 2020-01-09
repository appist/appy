package appy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type ViewEngineSuite struct {
	TestSuite
	asset      *Asset
	config     *Config
	logger     *Logger
	viewEngine *ViewEngine
	support    Supporter
}

func (s *ViewEngineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, _, _ = NewFakeLogger()
	s.asset = NewAsset(nil, map[string]string{
		"docker": "testdata/viewengine/.docker",
		"config": "testdata/viewengine/configs",
		"locale": "testdata/viewengine/pkg/locales",
		"view":   "testdata/viewengine/pkg/views",
		"web":    "testdata/viewengine/web",
	})
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)
}

func (s *ViewEngineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ViewEngineSuite) TestNewViewEngine() {
	s.NotNil(s.viewEngine)
	s.NotNil(s.viewEngine.Set)
}

func (s *ViewEngineSuite) TestSetGlobalFuncs() {
	originalAssetPath, found := s.viewEngine.LookupGlobal("assetPath")

	s.viewEngine.SetGlobalFuncs(map[string]interface{}{
		"test": func() string {
			return "test"
		},
		"assetPath": func() string {
			return "customAssetPath"
		},
	})

	assetPath, found := s.viewEngine.LookupGlobal("assetPath")
	s.NotEqual(originalAssetPath, assetPath)
	s.Equal(true, found)

	_, found = s.viewEngine.LookupGlobal("test")
	s.Equal(true, found)
}

func (s *ViewEngineSuite) TestAssetPathWithDebugBuild() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.Equal("/manifest.json", req.URL.String())
		w.Write([]byte(`{"/images/a.png":"/images/a.contenthash.png"}`))
	}))
	defer server.Close()

	s.Panics(func() { s.viewEngine.assetPath("/images/a.png") })

	s.viewEngine.httpClient = server.Client()
	s.viewEngine.manifestHostname = server.URL
	s.Equal("/images/a.contenthash.png", s.viewEngine.assetPath("/images/a.png"))
	s.Panics(func() { s.viewEngine.assetPath("/images/404.png") })

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.Equal("/manifest.json", req.URL.String())
		w.Write([]byte(`not a json`))
	}))

	s.viewEngine.httpClient = server.Client()
	s.viewEngine.manifestHostname = server.URL
	s.Panics(func() { s.viewEngine.assetPath("/images/a.png") })

	s.config.HTTPSSLEnabled = true
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.Equal("/manifest.json", req.URL.String())
		w.Write([]byte(`{"/images/a.png":"/images/a.contenthash.png"}`))
	}))

	s.viewEngine.httpClient = server.Client()
	s.viewEngine.manifestHostname = server.URL
	s.Equal("/images/a.contenthash.png", s.viewEngine.assetPath("/images/a.png"))
	s.Panics(func() { s.viewEngine.assetPath("/images/404.png") })
}

func (s *ViewEngineSuite) TestAssetPathWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	s.asset = NewAsset(nil, nil)
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)
	s.Panics(func() { s.viewEngine.assetPath("/images/a.png") })

	s.asset = NewAsset(http.Dir("testdata/viewengine"), map[string]string{
		"docker": "testdata/viewengine/.docker",
		"config": "testdata/viewengine/configs",
		"locale": "testdata/viewengine/pkg/locales",
		"view":   "testdata/viewengine/pkg/views",
		"web":    "testdata/viewengine/web",
	})
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)

	s.Equal("/images/a.contenthash.png", s.viewEngine.assetPath("/images/a.png"))
	s.Panics(func() { s.viewEngine.assetPath("/images/404.png") })
}

func TestViewEngineSuite(t *testing.T) {
	RunTestSuite(t, new(ViewEngineSuite))
}
