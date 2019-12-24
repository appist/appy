package support

import (
	"net/http"
	"testing"

	"github.com/appist/appy/test"
)

type ViewLoaderSuite struct {
	test.Suite
}

func (s *ViewLoaderSuite) SetupTest() {
}

func (s *ViewLoaderSuite) TearDownTest() {
}

func (s *ViewLoaderSuite) TestNewViewLoader() {
	assets := NewAssets(nil, "", nil)
	viewLoader := NewViewLoader(assets)
	s.NotNil(viewLoader)
	s.NotNil(viewLoader.assets)
}

func (s *ViewLoaderSuite) TestOpenWithDebugBuild() {
	layout := map[string]string{
		"docker": "testdata/.docker",
		"config": "testdata/configs",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}
	assets := NewAssets(layout, "", nil)
	viewLoader := NewViewLoader(assets)
	reader, err := viewLoader.Open("testdata/pkg/views/home/index.html")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *ViewLoaderSuite) TestOpenWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	assets := NewAssets(nil, "", http.Dir("./testdata"))
	viewLoader := NewViewLoader(assets)
	reader, err := viewLoader.Open("pkg/views/home/index.html")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *ViewLoaderSuite) TestExistsWithDebugBuild() {
	layout := map[string]string{
		"docker": "testdata/.docker",
		"config": "testdata/configs",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}
	assets := NewAssets(layout, "", nil)
	viewLoader := NewViewLoader(assets)
	filename, exists := viewLoader.Exists("home/index.html")
	s.Equal("testdata/pkg/views/home/index.html", filename)
	s.Equal(true, exists)

	filename, exists = viewLoader.Exists("home/about.html")
	s.Equal("", filename)
	s.Equal(false, exists)
}

func (s *ViewLoaderSuite) TestExistsWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	assets := NewAssets(nil, "", http.Dir("./testdata"))
	viewLoader := NewViewLoader(assets)
	filename, exists := viewLoader.Exists("home/index.html")
	s.Equal("pkg/views/home/index.html", filename)
	s.Equal(true, exists)

	filename, exists = viewLoader.Exists("home/about.html")
	s.Equal("", filename)
	s.Equal(false, exists)
}

func TestViewLoaderSuite(t *testing.T) {
	test.RunSuite(t, new(ViewLoaderSuite))
}
