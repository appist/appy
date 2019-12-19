package support

import (
	"net/http"
	"testing"

	"github.com/appist/appy/test"
)

type AssetsSuite struct {
	test.Suite
}

func (s *AssetsSuite) SetupTest() {
}

func (s *AssetsSuite) TearDownTest() {
}

func (s *AssetsSuite) TestNewAssetsDefaultValue() {
	assets := NewAssets(nil, "", nil)
	s.Equal(map[string]string{
		"docker": ".docker",
		"config": "pkg/config",
		"locale": "pkg/locales",
		"view":   "pkg/views",
		"web":    "web",
	}, assets.Layout())
	s.Equal(".ssr", assets.ssrRelease)
	s.Nil(assets.static)
}

func (s *AssetsSuite) TestNewAssetsOpenWithDebugBuild() {
	assets := NewAssets(nil, "", nil)
	reader, err := assets.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open pkg/config/.env.missing: no such file or directory")

	reader, err = assets.Open("testdata/pkg/config/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsSuite) TestNewAssetsOpenWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", http.Dir("testdata"))
	reader, err := assets.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open testdata/pkg/config/.env.missing: no such file or directory")

	reader, err = assets.Open("pkg/config/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsSuite) TestNewAssetsOpenWithReleaseBuildAndMissingStatic() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", nil)
	reader, err := assets.Open("pkg/config/.env.missing")
	s.Nil(reader)
	s.EqualError(err, ErrNoStaticAssets.Error())
}

func TestAssetsSuite(t *testing.T) {
	test.RunSuite(t, new(AssetsSuite))
}
