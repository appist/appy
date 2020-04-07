package view

import (
	"io"
	"io/ioutil"

	"github.com/appist/appy/support"
)

// Loader is used to load template content.
type Loader struct {
	asset *support.Asset
}

// NewLoader initializes the ViewLoader instance.
func NewLoader(asset *support.Asset) *Loader {
	return &Loader{
		asset: asset,
	}
}

// Open opens the underlying reader with template content.
func (l *Loader) Open(filename string) (io.ReadCloser, error) {
	reader, err := l.asset.Open(filename)

	return ioutil.NopCloser(reader), err
}

// Exists checks for template existence and returns full path.
func (l *Loader) Exists(filename string) (string, bool) {
	filename = l.asset.Layout().View() + "/" + filename
	_, err := l.asset.ReadFile(filename)

	if err != nil {
		return "", false
	}

	return filename, true
}
