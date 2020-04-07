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
	manifestPath = "/manifest.json"
)

// Engine renders the view template.
type Engine struct {
	*jet.Set
	asset            *support.Asset
	config           *support.Config
	logger           *support.Logger
	httpClient       *http.Client
	manifestHostname string
}

// NewEngine initializes the view engine instance.
func NewEngine(asset *support.Asset, config *support.Config, logger *support.Logger) *Engine {
	loader := jet.NewSetLoader(template.HTMLEscape, NewLoader(asset))

	return &Engine{
		loader,
		asset,
		config,
		logger,
		&http.Client{},
		"",
	}
}

// SetGlobalFuncs set up the global functions by combining built-in and application functions.
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
		e.AddGlobal(name, f)
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

		response, err := e.httpClient.Get(hostname + manifestPath)
		if err != nil {
			e.logger.Panic(err)
		}
		defer response.Body.Close()

		data, _ = ioutil.ReadAll(response.Body)
	} else {
		data, err = e.asset.ReadFile(manifestPath)

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
