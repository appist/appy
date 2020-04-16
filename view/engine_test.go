package view

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type engineSuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	logger *support.Logger
	engine *Engine
}

func (s *engineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.engine = NewEngine(s.asset, s.config, s.logger)
}

func (s *engineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *engineSuite) TestSetGlobalFuncs() {
	customAssetPath := func() string {
		return "customAssetPath"
	}

	s.engine.SetGlobalFuncs(map[string]interface{}{
		"test": func() string {
			return "test"
		},
		"assetPath": customAssetPath,
	})

	{
		assetPath, found := s.engine.HTMLSet().LookupGlobal("assetPath")
		s.NotEqual(&customAssetPath, &assetPath)
		s.Equal(true, found)

		_, found = s.engine.HTMLSet().LookupGlobal("test")
		s.Equal(true, found)
	}

	{
		assetPath, found := s.engine.TxtSet().LookupGlobal("assetPath")
		s.NotEqual(&customAssetPath, &assetPath)
		s.Equal(true, found)

		_, found = s.engine.TxtSet().LookupGlobal("test")
		s.Equal(true, found)
	}
}

func (s *engineSuite) TestAssetPathWithDebugBuild() {
	{
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			s.Equal("/manifest.json", req.URL.String())
			w.Write([]byte(`{"/images/a.png":"/images/a.contenthash.png"}`))
		}))
		defer server.Close()

		s.Panics(func() { s.engine.assetPath("/images/a.png") })
		s.engine.httpClient = server.Client()
		s.engine.manifestHostname = server.URL

		s.Equal("/images/a.contenthash.png", s.engine.assetPath("/images/a.png"))
		s.Panics(func() { s.engine.assetPath("/images/404.png") })
	}

	{
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			s.Equal("/manifest.json", req.URL.String())
			w.Write([]byte(`not a json`))
		}))
		s.engine.httpClient = server.Client()
		s.engine.manifestHostname = server.URL

		s.Panics(func() { s.engine.assetPath("/images/a.png") })
	}

	{
		s.config.HTTPSSLEnabled = true
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			s.Equal("/manifest.json", req.URL.String())
			w.Write([]byte(`{"/images/a.png":"/images/a.contenthash.png"}`))
		}))

		s.engine.httpClient = server.Client()
		s.engine.manifestHostname = server.URL

		s.Equal("/images/a.contenthash.png", s.engine.assetPath("/images/a.png"))
		s.Panics(func() { s.engine.assetPath("/images/404.png") })
	}
}

func (s *engineSuite) TestAssetPathWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	{
		s.asset = support.NewAsset(nil, "")
		s.config = support.NewConfig(s.asset, s.logger)
		s.engine = NewEngine(s.asset, s.config, s.logger)

		s.Panics(func() { s.engine.assetPath("/images/a.png") })
	}

	{
		s.asset = support.NewAsset(http.Dir("testdata/engine/asset_path_with_release_build"), "")
		s.config = support.NewConfig(s.asset, s.logger)
		s.engine = NewEngine(s.asset, s.config, s.logger)

		s.Equal("/images/a.contenthash.png", s.engine.assetPath("/images/a.png"))
		s.Panics(func() { s.engine.assetPath("/images/404.png") })
	}
}

func TestEngineSuite(t *testing.T) {
	test.Run(t, new(engineSuite))
}
