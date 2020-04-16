package view

import (
	"net/http"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type loaderSuite struct {
	test.Suite
}

func (s *loaderSuite) TestOpenWithDebugBuild() {
	asset := support.NewAsset(nil, "testdata/loader/open_with_debug_build")
	loader := NewLoader(asset)
	reader, err := loader.Open("pkg/views/home/index.html")

	s.Nil(err)
	s.NotNil(reader)
}

func (s *loaderSuite) TestOpenWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	asset := support.NewAsset(http.Dir("testdata/loader/open_with_release_build"), "")
	loader := NewLoader(asset)
	reader, err := loader.Open("pkg/views/home/index.html")

	s.Nil(err)
	s.NotNil(reader)
}

func (s *loaderSuite) TestExistsWithDebugBuild() {
	asset := support.NewAsset(nil, "testdata/loader/exists_with_debug_build")
	loader := NewLoader(asset)

	filename, exists := loader.Exists("home/index.html")
	s.Equal("pkg/views/home/index.html", filename)
	s.Equal(true, exists)

	filename, exists = loader.Exists("home/about.html")
	s.Equal("", filename)
	s.Equal(false, exists)
}

func (s *loaderSuite) TestExistsWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	asset := support.NewAsset(http.Dir("testdata/loader/exists_with_release_build"), "")
	loader := NewLoader(asset)

	filename, exists := loader.Exists("home/index.html")
	s.Equal("pkg/views/home/index.html", filename)
	s.Equal(true, exists)

	filename, exists = loader.Exists("home/about.html")
	s.Equal("", filename)
	s.Equal(false, exists)
}

func TestLoaderSuite(t *testing.T) {
	test.Run(t, new(loaderSuite))
}
