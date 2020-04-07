package support

import (
	"net/http"
	"testing"

	"github.com/appist/appy/test"
)

type AssetSuite struct {
	test.Suite
}

func (s *AssetSuite) TestOpsInDebugBuild() {
	asset := NewAsset(nil, "testdata/asset/ops_in_debug_build")

	_, err := asset.Open(asset.Layout().view)
	s.Equal("open testdata/asset/ops_in_debug_build/pkg/views: no such file or directory", err.Error())

	dirs, err := asset.ReadDir(asset.Layout().locale)
	s.Nil(err)
	s.Equal(3, len(dirs))

	_, err = asset.ReadDir(asset.Layout().view)
	s.Equal("open testdata/asset/ops_in_debug_build/pkg/views: no such file or directory", err.Error())

	data, err := asset.ReadFile(asset.Layout().config + "/.env.development")
	s.Nil(err)
	s.Equal("HTTP_HOST=0.0.0.0\n", string(data))

	_, err = asset.ReadFile(asset.Layout().view + "/index.html")
	s.Equal("open testdata/asset/ops_in_debug_build/pkg/views/index.html: no such file or directory", err.Error())
}

func (s *AssetSuite) TestOpsInReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	{
		asset := NewAsset(nil, "")

		_, err := asset.Open(asset.Layout().view)
		s.Equal(ErrNoEmbeddedAssets, err)

		_, err = asset.ReadDir(asset.Layout().locale)
		s.Equal(ErrNoEmbeddedAssets, err)

		_, err = asset.ReadFile(asset.Layout().config + "/.env.development")
		s.Equal(ErrNoEmbeddedAssets, err)
	}

	{
		asset := NewAsset(
			http.Dir("./testdata/asset/ops_in_release_build"),
			"",
		)

		_, err := asset.Open(asset.Layout().view)
		s.Equal("open testdata/asset/ops_in_release_build/pkg/views: no such file or directory", err.Error())

		dirs, err := asset.ReadDir(asset.Layout().locale)
		s.Nil(err)
		s.Equal(3, len(dirs))

		_, err = asset.ReadDir(asset.Layout().view)
		s.Equal("open testdata/asset/ops_in_release_build/pkg/views: no such file or directory", err.Error())

		data, err := asset.ReadFile(asset.Layout().config + "/.env.development")
		s.Nil(err)
		s.Equal("HTTP_HOST=0.0.0.0\n", string(data))

		_, err = asset.ReadFile(asset.Layout().view + "/index.html")
		s.Equal("open testdata/asset/ops_in_release_build/pkg/views/index.html: no such file or directory", err.Error())
	}
}

func TestAssetSuite(t *testing.T) {
	test.Run(t, new(AssetSuite))
}
