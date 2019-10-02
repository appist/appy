package http

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	ssrRoot          = "app"
	ssrView          = "views"
	ssrLocale        = "locales"
	reservedViewDirs = []string{"layouts", "shared"}
)

func init() {
	if support.Build == "release" {
		ssrRoot = ".ssr"
	}
}

// InitSSRView loads all the view files for HTML rendering.
func (s *ServerT) InitSSRView() error {
	viewDir := ssrRoot + "/" + ssrView

	// Try getting all the view files from `app/views`, but fallback to `assets` http.FileSystem.
	viewDirs, err := ioutil.ReadDir(viewDir)
	if err != nil {
		dir, err := s.Assets.Open("/" + viewDir)
		if err != nil {
			return err
		}

		viewDirs, err = dir.Readdir(-1)
	}

	// for _, fi := range reservedViewDirs {

	// }

	for _, fi := range viewDirs {
		if fi.IsDir() && support.Contains(reservedViewDirs, fi.Name()) == true {
			continue
		}
	}

	return nil
}

// InitSSRLocale loads all the locale files to initialize the I18n bundle.
func (s *ServerT) InitSSRLocale() error {
	i18nBundle := i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	localeDir := ssrRoot + "/" + ssrLocale

	// Try getting all the locale files from `app/locales`, but fallback to `assets` http.FileSystem.
	localeFiles, err := ioutil.ReadDir(localeDir)
	if err != nil {
		file, err := s.Assets.Open("/" + localeDir)
		if err != nil {
			return err
		}

		localeFiles, err = file.Readdir(-1)
	}

	for _, localeFile := range localeFiles {
		localeFn := localeFile.Name()
		data, err := ioutil.ReadFile(localeDir + "/" + localeFn)
		if err != nil && os.IsNotExist(err) {
			file, err := s.Assets.Open("/" + localeDir + "/" + localeFn)
			if err != nil {
				return err
			}

			data, err = ioutil.ReadAll(file)
			if err != nil {
				return err
			}
		}

		i18nBundle.MustParseMessageFileBytes(data, localeFn)
	}

	s.Router.Use(middleware.I18n(i18nBundle))
	return nil
}
