package appy

import (
	"io"
	"io/ioutil"
)

// ViewLoader is used to load template content.
type ViewLoader struct {
	asset *Asset
}

// NewViewLoader initializes the ViewLoader instance.
func NewViewLoader(asset *Asset) *ViewLoader {
	return &ViewLoader{
		asset: asset,
	}
}

// Open opens the underlying reader with template content.
func (vl *ViewLoader) Open(filename string) (io.ReadCloser, error) {
	reader, err := vl.asset.Open(filename)

	return ioutil.NopCloser(reader), err
}

// Exists checks for template existence and returns full path.
func (vl *ViewLoader) Exists(filename string) (string, bool) {
	filename = vl.asset.Layout()["view"] + "/" + filename
	_, err := vl.asset.ReadFile(filename)

	if err != nil {
		return "", false
	}

	return filename, true
}
