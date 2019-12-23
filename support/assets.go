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

// ReadDir reads the directory named by dirname and returns a list of file/directory entries.
func (m Assets) ReadDir(dirname string) ([]os.FileInfo, error) {
	var (
		fis []os.FileInfo
		err error
	)

	if IsDebugBuild() {
		fis, err = ioutil.ReadDir(dirname)
	} else {
		dirname = m.normalizedPath(dirname)
		reader, err := m.static.Open(dirname)

		if err != nil {
			return nil, err
		}

		fis, err = reader.Readdir(-1)
	}

	return fis, err
}

// ReadFile reads the file named by filename and returns the contents.
func (m Assets) ReadFile(filename string) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	filename = m.normalizedPath(filename)

	if IsDebugBuild() {
		data, err = ioutil.ReadFile(filename)
	} else {
		file, err := m.static.Open(filename)

		if err != nil {
			return nil, err
		}

		data, err = ioutil.ReadAll(file)
	}

	return data, err
}

func (m Assets) normalizedPath(path string) string {
	if IsDebugBuild() {
		return path
	}

	return "/" + m.ssrRelease + "/" + path
}
