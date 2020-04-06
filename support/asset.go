package support

import "net/http"

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
