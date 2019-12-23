package support

import (
	"io"
	"io/ioutil"
)

type (
	// ViewLoader is used to load template content.
	ViewLoader struct {
		assets *Assets
	}
)

// NewViewLoader initializes the ViewLoader instance.
func NewViewLoader(assets *Assets) *ViewLoader {
	return &ViewLoader{
		assets: assets,
	}
}

// Open opens the underlying reader with template content.
func (v *ViewLoader) Open(filename string) (io.ReadCloser, error) {
	reader, err := v.assets.Open(filename)

	return ioutil.NopCloser(reader), err
}

// Exists checks for template existence and returns full path.
func (v *ViewLoader) Exists(filename string) (string, bool) {
	filename = v.assets.Layout()["view"] + "/" + filename
	_, err := v.assets.ReadFile(filename)

	if err != nil {
		return "", false
	}

	return filename, true
}
