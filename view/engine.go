package view

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/support"
)

const (
	assetsManifestPath = "/assets-manifest.json"
)

// Engine renders the view template.
type Engine struct {
	htmlSet          *jet.Set
	txtSet           *jet.Set
	asset            *support.Asset
	config           *support.Config
	logger           *support.Logger
	httpClient       *http.Client
	manifestHostname string
}

// NewEngine initializes the view engine instance.
func NewEngine(asset *support.Asset, config *support.Config, logger *support.Logger) *Engine {
	htmlSet := jet.NewSetLoader(template.HTMLEscape, NewLoader(asset))
	txtSet := jet.NewSetLoader(nil, NewLoader(asset))

	return &Engine{
		htmlSet,
		txtSet,
		asset,
		config,
		logger,
		&http.Client{},
		"",
	}
}

// HTMLSet returns the template set in which the contents are escaped by
// template.HTMLEscape.
func (e *Engine) HTMLSet() *jet.Set {
	return e.htmlSet
}

// TxtSet returns the template set in which the contents are plain text
// without being escaped.
func (e *Engine) TxtSet() *jet.Set {
	return e.txtSet
}

// SetGlobalFuncs set up the global functions by combining built-in and
// application functions.
func (e *Engine) SetGlobalFuncs(viewFuncs map[string]interface{}) {
	funcs := map[string]interface{}{
		"assetPath": e.assetPath,
	}

	for viewKey, viewFunc := range viewFuncs {
		if _, exists := funcs[viewKey]; exists {
			continue
		}

		funcs[viewKey] = viewFunc
	}

	for name, f := range funcs {
		e.htmlSet.AddGlobal(name, f)
		e.txtSet.AddGlobal(name, f)
	}
}

func (e *Engine) assetPath(path string) string {
	var (
		data     []byte
		err      error
		manifest map[string]interface{}
	)

	if support.IsDebugBuild() {
		scheme := "http://"
		port, _ := strconv.Atoi(e.config.HTTPPort)
		if e.config.HTTPSSLEnabled {
			scheme = "https://"
			port, _ = strconv.Atoi(e.config.HTTPSSLPort)
		}

		hostname := scheme + e.config.HTTPHost + ":" + strconv.Itoa(port+1)
		if e.manifestHostname != "" {
			hostname = e.manifestHostname
		}

		req, err := http.NewRequest("GET", hostname+assetsManifestPath, nil)
		if err != nil {
			e.logger.Panic(err)
		}

		req.Header.Set("Accept", "application/json")
		response, err := e.httpClient.Do(req)
		if err != nil {
			e.logger.Panic(err)
		}
		defer response.Body.Close()

		data, _ = ioutil.ReadAll(response.Body)
	} else {
		data, err = e.asset.ReadFile(assetsManifestPath)

		if err != nil {
			e.logger.Panic(err)
		}
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		e.logger.Panic(err)
	}

	if _, exists := manifest[path]; !exists {
		e.logger.Panic(os.ErrNotExist)
	}

	return e.config.AssetHost + manifest[path].(string)
}
