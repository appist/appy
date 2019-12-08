package support

import (
	"io"
	"net/http"
	"os"
)

type (
	// AssetsMngr manages all the assets for both CSR/SSR based on the current build type, i.e. debug or release.
	AssetsMngr struct {
		layout     map[string]string
		ssrRelease string
		static     http.FileSystem
	}
)

// NewAssetsMngr initializes the assets manager instance.
func NewAssetsMngr(layout map[string]string, ssrRelease string, static http.FileSystem) *AssetsMngr {
	if layout == nil {
		layout = map[string]string{
			"docker": ".docker",
			"config": "pkg/config",
			"locale": "pkg/locales",
			"view":   "pkg/views",
			"web":    "web",
		}
	}

	if ssrRelease == "" {
		ssrRelease = ".ssr"
	}

	return &AssetsMngr{
		layout:     layout,
		ssrRelease: ssrRelease,
		static:     static,
	}
}

// Layout returns the appy's project layout.
func (m AssetsMngr) Layout() map[string]string {
	return m.layout
}

// Open opens the named file for reading. If the current build type is debug, reads from the filesystem. Otherwise, it
// reads from the embedded static assets which is a virtual file system.
func (m AssetsMngr) Open(path string) (io.Reader, error) {
	var (
		reader io.Reader
		err    error
	)

	if IsDebugBuild() {
		reader, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	} else {
		if m.static == nil {
			return nil, ErrNoStaticAssets
		}

		reader, err = m.static.Open(path)
		if err != nil {
			return nil, err
		}
	}

	return reader, nil
}
