package appy_test

import (
	"net/http"
	"testing"

	"github.com/appist/appy"
)

type AssetSuite struct {
	appy.TestSuite
}

func (s *AssetSuite) TestNewAsset() {
	asset := appy.NewAsset(nil, nil, "")

	s.Equal("configs", asset.Layout()["config"])
	s.Equal(".docker", asset.Layout()["docker"])
	s.Equal("pkg/locales", asset.Layout()["locale"])
	s.Equal("pkg/views", asset.Layout()["view"])
	s.Equal("web", asset.Layout()["web"])

	asset = appy.NewAsset(nil, map[string]string{
		"config": "testdata/configs",
		"docker": "testdata/.docker",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}, "")

	s.Equal("testdata/configs", asset.Layout()["config"])
	s.Equal("testdata/.docker", asset.Layout()["docker"])
	s.Equal("testdata/pkg/locales", asset.Layout()["locale"])
	s.Equal("testdata/pkg/views", asset.Layout()["view"])
	s.Equal("testdata/web", asset.Layout()["web"])
}

func (s *AssetSuite) TestNewAssetWithModuleRoot() {
	asset := appy.NewAsset(nil, nil, "appist/appist")

	s.Equal("appist/appist/configs", asset.Layout()["config"])
	s.Equal("appist/appist/.docker", asset.Layout()["docker"])
	s.Equal("appist/appist/pkg/locales", asset.Layout()["locale"])
	s.Equal("appist/appist/pkg/views", asset.Layout()["view"])
	s.Equal("appist/appist/web", asset.Layout()["web"])
}

func (s *AssetSuite) TestOpen() {
	asset := appy.NewAsset(nil, map[string]string{
		"config": "testdata/asset/open/configs",
		"docker": "testdata/asset/open/.docker",
		"locale": "testdata/asset/open/pkg/locales",
		"view":   "testdata/asset/open/pkg/views",
		"web":    "testdata/asset/open/web",
	}, "")
	_, err := asset.Open("testdata/asset/open/configs/.env.development")
	s.EqualError(err, "open testdata/asset/open/configs/.env.development: no such file or directory")

	_, err = asset.Open("testdata/asset/open/pkg/locales/en.yml")
	s.NoError(err)

	appy.Build = appy.ReleaseBuild
	defer func() {
		appy.Build = appy.DebugBuild
	}()

	asset = appy.NewAsset(nil, nil, "")
	_, err = asset.Open("configs/.env.development")
	s.EqualError(err, appy.ErrNoEmbeddedAssets.Error())

	asset = appy.NewAsset(http.Dir("testdata/asset/open"), nil, "")
	_, err = asset.Open("configs/.env.development")
	s.EqualError(err, "open testdata/asset/open/configs/.env.development: no such file or directory")

	_, err = asset.Open("pkg/locales/en.yml")
	s.NoError(err)
}

func (s *AssetSuite) TestReadDir() {
	asset := appy.NewAsset(nil, map[string]string{
		"config": "testdata/asset/open/configs",
		"docker": "testdata/asset/open/.docker",
		"locale": "testdata/asset/open/pkg/locales",
		"view":   "testdata/asset/open/pkg/views",
		"web":    "testdata/asset/open/web",
	}, "")
	_, err := asset.ReadDir("testdata/asset/missing")
	s.EqualError(err, "open testdata/asset/missing: no such file or directory")

	_, err = asset.ReadDir("testdata/asset/readdir")
	s.NoError(err)

	appy.Build = appy.ReleaseBuild
	defer func() {
		appy.Build = appy.DebugBuild
	}()

	asset = appy.NewAsset(nil, nil, "")
	_, err = asset.ReadDir("configs")
	s.EqualError(err, appy.ErrNoEmbeddedAssets.Error())

	asset = appy.NewAsset(http.Dir("testdata/asset/readdir"), nil, "")
	_, err = asset.ReadDir("configs")
	s.EqualError(err, "open testdata/asset/readdir/configs: no such file or directory")

	fis, err := asset.ReadDir("pkg")
	s.NoError(err)
	s.Equal(2, len(fis))
}

func (s *AssetSuite) TestReadFile() {
	asset := appy.NewAsset(nil, map[string]string{
		"config": "testdata/asset/open/configs",
		"docker": "testdata/asset/open/.docker",
		"locale": "testdata/asset/open/pkg/locales",
		"view":   "testdata/asset/open/pkg/views",
		"web":    "testdata/asset/open/web",
	}, "")
	_, err := asset.ReadFile("testdata/asset/missing/pkg/locales/en.yml")
	s.EqualError(err, "open testdata/asset/missing/pkg/locales/en.yml: no such file or directory")

	_, err = asset.ReadFile("testdata/asset/readfile/pkg/locales/en.yml")
	s.NoError(err)

	appy.Build = appy.ReleaseBuild
	defer func() {
		appy.Build = appy.DebugBuild
	}()

	asset = appy.NewAsset(nil, nil, "")
	_, err = asset.ReadFile("pkg/locales/en.yml")
	s.EqualError(err, appy.ErrNoEmbeddedAssets.Error())

	asset = appy.NewAsset(http.Dir("testdata/asset/readfile"), nil, "")
	_, err = asset.ReadFile("pkg/locales/en-US.yml")
	s.EqualError(err, "open testdata/asset/readfile/pkg/locales/en-US.yml: no such file or directory")

	data, err := asset.ReadFile("pkg/locales/en.yml")
	s.NoError(err)
	s.Equal("test: Test\n", string(data))
}

func TestAssetSuite(t *testing.T) {
	appy.RunTestSuite(t, new(AssetSuite))
}
