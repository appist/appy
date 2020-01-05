package support

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/CloudyKit/jet"
)

type (
	// ViewEngine renders the view template.
	ViewEngine struct {
		*jet.Set
		assets           *Assets
		config           *Config
		logger           *Logger
		httpClient       *http.Client
		manifestHostname string
	}
)

// NewViewEngine initializes the view engine instance.
func NewViewEngine(assets *Assets, config *Config, logger *Logger) *ViewEngine {
	return &ViewEngine{
		jet.NewSetLoader(template.HTMLEscape, NewViewLoader(assets)),
		assets,
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

	if funcs != nil {
		for name, f := range funcs {
			ve.AddGlobal(name, f)
		}
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
		data, err = ve.assets.ReadFile(manifestPath, false)

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

	return manifest[path].(string)
}
