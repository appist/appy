package appy

import (
	"net/http"
	"testing"
)

type ViewLoaderSuite struct {
	TestSuite
}

func (s *ViewLoaderSuite) SetupTest() {
}

func (s *ViewLoaderSuite) TearDownTest() {
}

func (s *ViewLoaderSuite) TestNewViewLoader() {
	asset := NewAsset(nil, nil, "")
	viewLoader := NewViewLoader(asset)
	s.NotNil(viewLoader)
	s.NotNil(viewLoader.asset)
}

func (s *ViewLoaderSuite) TestOpenWithDebugBuild() {
	asset := NewAsset(nil, map[string]string{
		"docker": "testdata/viewloader/.docker",
		"config": "testdata/viewloader/configs",
		"locale": "testdata/viewloader/pkg/locales",
		"view":   "testdata/viewloader/pkg/views",
		"web":    "testdata/viewloader/web",
	}, "")
	viewLoader := NewViewLoader(asset)
	reader, err := viewLoader.Open("testdata/viewloader/pkg/views/home/index.html")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *ViewLoaderSuite) TestOpenWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() {
		Build = DebugBuild
	}()

	asset := NewAsset(http.Dir("testdata/viewloader"), nil, "")
	viewLoader := NewViewLoader(asset)
	reader, err := viewLoader.Open("pkg/views/home/index.html")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *ViewLoaderSuite) TestExistsWithDebugBuild() {
	asset := NewAsset(nil, map[string]string{
		"docker": "testdata/viewloader/.docker",
		"config": "testdata/viewloader/configs",
		"locale": "testdata/viewloader/pkg/locales",
		"view":   "testdata/viewloader/pkg/views",
		"web":    "testdata/viewloader/web",
	}, "")
	viewLoader := NewViewLoader(asset)
	filename, exists := viewLoader.Exists("home/index.html")
	s.Equal("testdata/viewloader/pkg/views/home/index.html", filename)
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

	asset := NewAsset(http.Dir("testdata/viewloader"), nil, "")
	viewLoader := NewViewLoader(asset)
	filename, exists := viewLoader.Exists("home/index.html")
	s.Equal("pkg/views/home/index.html", filename)
	s.Equal(true, exists)

	filename, exists = viewLoader.Exists("home/about.html")
	s.Equal("", filename)
	s.Equal(false, exists)
}

func TestViewLoaderSuite(t *testing.T) {
	RunTestSuite(t, new(ViewLoaderSuite))
}
