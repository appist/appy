package support

import (
	"io"
	"net/http"
	"os"
)

type (
	// Assets manages all the assets for both CSR/SSR based on the current build type, i.e. debug or release.
	Assets struct {
		layout     map[string]string
		ssrRelease string
		static     http.FileSystem
	}
)

// NewAssets initializes the assets instance.
func NewAssets(layout map[string]string, ssrRelease string, static http.FileSystem) *Assets {
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

	return &Assets{
		layout:     layout,
		ssrRelease: ssrRelease,
		static:     static,
	}
}

// Layout returns the appy's project layout.
func (m Assets) Layout() map[string]string {
	return m.layout
}

// Open opens the named file for reading. If the current build type is debug, reads from the filesystem. Otherwise, it
// reads from the embedded static assets which is a virtual file system.
func (m Assets) Open(path string) (io.Reader, error) {
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
