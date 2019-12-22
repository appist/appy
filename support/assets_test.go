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
		"config": "configs",
		"locale": "pkg/locales",
		"view":   "pkg/views",
		"web":    "web",
	}, assets.Layout())
	s.Equal(".ssr", assets.ssrRelease)
	s.Nil(assets.static)
}

func (s *AssetsSuite) TestNewAssetsOpenWithDebugBuild() {
	assets := NewAssets(nil, "", nil)
	reader, err := assets.Open("configs/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open configs/.env.missing: no such file or directory")

	reader, err = assets.Open("testdata/configs/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsSuite) TestNewAssetsOpenWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", http.Dir("testdata"))
	reader, err := assets.Open("configs/.env.missing")
	s.Nil(reader)
	s.EqualError(err, "open testdata/configs/.env.missing: no such file or directory")

	reader, err = assets.Open("configs/.env.development")
	s.NotNil(reader)
	s.Nil(err)
}

func (s *AssetsSuite) TestNewAssetsOpenWithReleaseBuildAndMissingStatic() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", nil)
	reader, err := assets.Open("configs/.env.missing")
	s.Nil(reader)
	s.EqualError(err, ErrNoStaticAssets.Error())
}

func (s *AssetsSuite) TestNewAssetsReadDirWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", http.Dir("testdata"))
	_, err := assets.ReadDir("./foo")
	s.NotNil(err)

	fis, err := assets.ReadDir("./pkg/locales")
	s.Nil(err)
	s.Equal("en.yml", fis[0].Name())
}

func (s *AssetsSuite) TestNewAssetsReadFileWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	assets := NewAssets(nil, "", http.Dir("testdata"))
	_, err := assets.ReadFile("./foo")
	s.NotNil(err)

	_, err = assets.ReadFile("./pkg/locales/zh-TW.yml")
	s.NotNil(err)

	data, err := assets.ReadFile("./pkg/locales/en.yml")
	s.Nil(err)
	s.Equal("title: Test\n", string(data))
}

func TestAssetsSuite(t *testing.T) {
	test.RunSuite(t, new(AssetsSuite))
}
