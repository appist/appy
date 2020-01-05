package support

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type ViewEngineSuite struct {
	test.Suite
	assets     *Assets
	config     *Config
	logger     *Logger
	viewEngine *ViewEngine
}

func (s *ViewEngineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewFakeLogger()
	s.assets = NewAssets(map[string]string{
		"docker": "testdata/.docker",
		"config": "testdata/configs",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}, "", nil)
	s.config = NewConfig(s.assets, s.logger)
	s.viewEngine = NewViewEngine(s.assets, s.config, s.logger)
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
	defer server.Close()

	s.viewEngine.httpClient = server.Client()
	s.viewEngine.manifestHostname = server.URL
	s.Panics(func() { s.viewEngine.assetPath("/images/a.png") })
}

func (s *ViewEngineSuite) TestAssetPathWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	s.assets = NewAssets(nil, "", http.Dir("testdata/.ssr"))
	s.config = NewConfig(s.assets, s.logger)
	s.viewEngine = NewViewEngine(s.assets, s.config, s.logger)
	s.Panics(func() { s.viewEngine.assetPath("/images/a.png") })

	s.assets = NewAssets(map[string]string{
		"docker": "testdata/.docker",
		"config": "testdata/configs",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}, "", http.Dir("testdata"))
	s.config = NewConfig(s.assets, s.logger)
	s.viewEngine = NewViewEngine(s.assets, s.config, s.logger)

	s.Equal("/images/a.contenthash.png", s.viewEngine.assetPath("/images/a.png"))
	s.Panics(func() { s.viewEngine.assetPath("/images/404.png") })
}

func TestViewEngineSuite(t *testing.T) {
	test.RunSuite(t, new(ViewEngineSuite))
}
