package appy

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/CloudyKit/jet"
)

// ViewEngine renders the view template.
type ViewEngine struct {
	*jet.Set
	asset            *Asset
	config           *Config
	logger           *Logger
	httpClient       *http.Client
	manifestHostname string
}

// NewViewEngine initializes the view engine instance.
func NewViewEngine(asset *Asset, config *Config, logger *Logger) *ViewEngine {
	viewLoader := jet.NewSetLoader(template.HTMLEscape, NewViewLoader(asset))

	return &ViewEngine{
		viewLoader,
		asset,
		config,
		logger,
		&http.Client{},
		"",
	}
}

// SetGlobalFuncs set up the global functions by combining built-in and application functions.
func (ve *ViewEngine) SetGlobalFuncs(viewFuncs map[string]interface{}) {
	funcs := map[string]interface{}{
		"assetPath": ve.assetPath,
	}

	for viewKey, viewFunc := range viewFuncs {
		if _, exists := funcs[viewKey]; exists {
			continue
		}

		funcs[viewKey] = viewFunc
	}

	for name, f := range funcs {
		ve.AddGlobal(name, f)
	}
}

func (ve *ViewEngine) assetPath(path string) string {
	var (
		data     []byte
		err      error
		manifest map[string]interface{}
	)

	manifestPath := "/manifest.json"

	if IsDebugBuild() {
		scheme := "http://"
		port, _ := strconv.Atoi(ve.config.HTTPPort)
		if ve.config.HTTPSSLEnabled {
			scheme = "https://"
			port, _ = strconv.Atoi(ve.config.HTTPSSLPort)
		}

		hostname := scheme + ve.config.HTTPHost + ":" + strconv.Itoa(port+1)
		if ve.manifestHostname != "" {
			hostname = ve.manifestHostname
		}

		response, err := ve.httpClient.Get(hostname + manifestPath)
		if err != nil {
			ve.logger.Panic(err)
		}
		defer response.Body.Close()

		data, _ = ioutil.ReadAll(response.Body)
	} else {
		data, err = ve.asset.ReadFile(manifestPath)

		if err != nil {
			ve.logger.Panic(err)
		}
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		ve.logger.Panic(err)
	}

	if _, exists := manifest[path]; !exists {
		ve.logger.Panic(os.ErrNotExist)
	}

	return ve.config.AssetHost + manifest[path].(string)
}
