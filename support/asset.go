package support

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type (
	// AssetManager implements all methods for Asset.
	AssetManager interface {
		Layout() *AssetLayout
		Open(path string) (io.Reader, error)
		ReadDir(dir string) ([]os.FileInfo, error)
		ReadFile(filename string) ([]byte, error)
	}

	// Asset manages the application assets.
	Asset struct {
		embedded http.FileSystem
		layout   *AssetLayout
	}
)

// NewAsset initializes the assets instance.
func NewAsset(embedded http.FileSystem, root string) *Asset {
	asset := &Asset{
		embedded: embedded,
		layout: &AssetLayout{
			config: "configs",
			docker: ".docker",
			locale: "pkg/locales",
			root:   root,
			view:   "pkg/views",
			web:    "web",
		},
	}

	return asset
}

// Embedded returns the embedded asset.
func (a *Asset) Embedded() http.FileSystem {
	return a.embedded
}

// Layout keeps the path for project components.
func (a *Asset) Layout() *AssetLayout {
	return a.layout
}

// Open opens the named file for reading. If the current build type is debug,
// reads from the filesystem. Otherwise, it reads from the embedded static
// assets which is a virtual file system.
func (a *Asset) Open(path string) (io.Reader, error) {
	if IsDebugBuild() {
		return os.Open(a.Layout().root + "/" + path)
	}

	if a.embedded == nil {
		return nil, ErrNoEmbeddedAssets
	}

	return a.embedded.Open(path)
}

// ReadDir returns a list of file/directory entries in the directory.
func (a *Asset) ReadDir(dir string) ([]os.FileInfo, error) {
	if IsDebugBuild() {
		return ioutil.ReadDir(a.Layout().root + "/" + dir)
	}

	if a.embedded == nil {
		return nil, ErrNoEmbeddedAssets
	}

	reader, err := a.embedded.Open(dir)
	if err != nil {
		return nil, err
	}

	return reader.Readdir(-1)
}

// ReadFile returns the content of the filename.
func (a *Asset) ReadFile(filename string) ([]byte, error) {
	if IsDebugBuild() {
		return ioutil.ReadFile(a.Layout().root + "/" + filename)
	}

	if a.embedded == nil {
		return nil, ErrNoEmbeddedAssets
	}

	file, err := a.embedded.Open(filename)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(file)
}

// AssetLayout manages the path for project components.
type AssetLayout struct {
	config, docker, locale, root, view, web string
}

// Config returns the path that stores the configuration.
func (l *AssetLayout) Config() string {
	return l.config
}

// Docker returns the path that stores the docker related files.
func (l *AssetLayout) Docker() string {
	return l.docker
}

// Locale returns the path that stores the server-side rendering locales.
func (l *AssetLayout) Locale() string {
	return l.locale
}

// Root returns the project root path.
func (l *AssetLayout) Root() string {
	return l.root
}

// View returns the path that stores the server-side rendering views.
func (l *AssetLayout) View() string {
	return l.view
}

// Web returns the path that stores the client-side rendering web app.
func (l *AssetLayout) Web() string {
	return l.web
}
