package appy

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// Asset manages the application assets.
type Asset struct {
	embedded   http.FileSystem
	layout     map[string]string
	moduleRoot string
}

// NewAsset initializes the assets instance.
func NewAsset(embedded http.FileSystem, layout map[string]string, moduleRoot string) *Asset {
	asset := &Asset{
		embedded: embedded,
		layout: map[string]string{
			"config": "configs",
			"docker": ".docker",
			"locale": "pkg/locales",
			"view":   "pkg/views",
			"web":    "web",
		},
		moduleRoot: moduleRoot,
	}

	if layout != nil {
		asset.layout = layout
	}

	return asset
}

// Layout returns the appy's project layout.
func (a *Asset) Layout() map[string]string {
	if IsDebugBuild() {
		if a.moduleRoot != "" {
			layout := map[string]string{}

			for key, val := range a.layout {
				layout[key] = a.moduleRoot + "/" + val
			}

			return layout
		}
	}

	return a.layout
}

// Open opens the named file for reading. If the current build type is debug, reads from the filesystem. Otherwise, it
// reads from the embedded static assets which is a virtual file system.
func (a *Asset) Open(path string) (io.Reader, error) {
	if IsDebugBuild() {
		return os.Open(path)
	}

	if a.embedded == nil {
		return nil, ErrNoEmbeddedAssets
	}

	return a.embedded.Open(path)
}

// ReadDir returns a list of file/directory entries in the directory.
func (a *Asset) ReadDir(dir string) ([]os.FileInfo, error) {
	if IsDebugBuild() {
		return ioutil.ReadDir(dir)
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
		return ioutil.ReadFile(filename)
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
