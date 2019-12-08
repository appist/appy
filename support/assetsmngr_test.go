package support

import (
	"net/http"
	"testing"

	"github.com/appist/appy/test"
)

type AssetsMngrSuite struct {
	test.Suite
}

func (s *AssetsMngrSuite) SetupTest() {
}

func (s *AssetsMngrSuite) TearDownTest() {
}

func (s *AssetsMngrSuite) TestNewAssetsMngrDefaultValue() {
	assetsMngr := NewAssetsMngr(nil, "", nil)
	s.Equal(map[string]string{
		"docker": ".docker",
		"config": "pkg/config",
		"locale": "pkg/locales",
		"view":   "pkg/views",
		"web":    "web",
	}, assetsMngr.Layout())
	s.Equal(".ssr", assetsMngr.ssrRelease)
	s.Nil(assetsMngr.static)
}

func (s *AssetsMngrSuite) TestNewAssetsMngrOpenWithDebugBuild() {
	assetsMngr := NewAssetsMngr(nil, "", nil)
	reader, err := assetsMngr.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open pkg/config/.env.missing: no such file or directory")

	reader, err = assetsMngr.Open("testdata/pkg/config/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsMngrSuite) TestNewAssetsMngrOpenWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assetsMngr := NewAssetsMngr(nil, "", http.Dir("testdata"))
	reader, err := assetsMngr.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open testdata/pkg/config/.env.missing: no such file or directory")

	reader, err = assetsMngr.Open("pkg/config/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsMngrSuite) TestNewAssetsMngrOpenWithReleaseBuildAndMissingStatic() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assetsMngr := NewAssetsMngr(nil, "", nil)
	reader, err := assetsMngr.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, ErrNoStaticAssets.Error())
}

func TestAssetsMngrSuite(t *testing.T) {
	test.RunSuite(t, new(AssetsMngrSuite))
}
