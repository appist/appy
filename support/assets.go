package support

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type (
	// Assets manages all the assets for both CSR/SSR based on the current build type, i.e. debug or release.
	Assets struct {
		layout     map[string]string
		ssrRelease string
		static     http.FileSystem
		viewLoader *ViewLoader
	}
)

// NewAssets initializes the assets instance.
func NewAssets(layout map[string]string, ssrRelease string, static http.FileSystem) *Assets {
	if layout == nil {
		layout = map[string]string{
			"docker": ".docker",
			"config": "configs",
			"locale": "pkg/locales",
			"view":   "pkg/views",
			"web":    "web",
		}
	}

	if ssrRelease == "" {
		ssrRelease = ".ssr"
	}

	assets := &Assets{
		layout:     layout,
		ssrRelease: ssrRelease,
		static:     static,
	}
	assets.viewLoader = NewViewLoader(assets)

	return assets
}

// Layout returns the appy's project layout.
func (a Assets) Layout() map[string]string {
	return a.layout
}

// Open opens the named file for reading. If the current build type is debug, reads from the filesystem. Otherwise, it
// reads from the embedded static assets which is a virtual file system.
func (a Assets) Open(path string) (io.Reader, error) {
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
		if a.static == nil {
			return nil, ErrNoStaticAssets
		}

		reader, err = a.static.Open(path)

		if err != nil {
			return nil, err
		}
	}

	return reader, nil
}

// ReadDir reads the directory named by dirname and returns a list of file/directory entries.
func (a Assets) ReadDir(dirname string) ([]os.FileInfo, error) {
	if IsDebugBuild() {
		return ioutil.ReadDir(dirname)
	}

	dirname = a.normalizedPath(dirname)
	reader, err := a.static.Open(dirname)

	if err != nil {
		return nil, err
	}

	return reader.Readdir(-1)
}

// ReadFile reads the file named by filename and returns the contents.
func (a Assets) ReadFile(filename string) ([]byte, error) {
	filename = a.normalizedPath(filename)

	if IsDebugBuild() {
		return ioutil.ReadFile(filename)
	}

	file, err := a.static.Open(filename)

	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(file)
}

// SSRRelease returns the SSR release path.
func (a Assets) SSRRelease() string {
	return a.ssrRelease
}

// Static returns the application static assets.
func (a Assets) Static() http.FileSystem {
	return a.static
}

func (a Assets) normalizedPath(path string) string {
	if IsDebugBuild() {
		return path
	}

	return "/" + a.ssrRelease + "/" + path
}
