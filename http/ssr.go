package http

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"appist/appy/middleware"
	"appist/appy/support"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	// LocaleDir is used to store the server-side translation files.
	LocaleDir = "locale"

	// ViewDir is used to store the server-side view templates.
	ViewDir = "view"

	// SSRAssetsDir is used to specify which folder to organise the SSR files like locales/views in the `assets` folder
	// which will be used to generate `main_assets.go` to be bundled in the single binary.
	SSRAssetsDir = ".ssr/"
)

// AddView adds a view template to the HTML renderer.
func (s *ServerT) AddView(name, layout string, templates []string) {
	tpls := []string{}
	tpl, err := s.viewTpl(ViewDir + "/" + layout)
	if err != nil {
		support.Logger.Error(err)
	}
	tpls = append(tpls, string(tpl))

	for _, t := range templates {
		tpl, err := s.viewTpl(ViewDir + "/" + t)

		if err != nil {
			support.Logger.Error(err)
		}

		tpls = append(tpls, string(tpl))
	}

	s.htmlRenderer.AddFromStringsFuncs(name, s.funcMap, tpls...)
}

func (s *ServerT) viewTpl(path string) ([]byte, error) {
	data, err := ioutil.ReadFile("app/" + path)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := s.assets.Open("/" + SSRAssetsDir + path)
			if err != nil {
				return nil, err
			}

			data, err := ioutil.ReadAll(file)
			if err != nil {
				return nil, err
			}

			return data, nil
		}

		return nil, err
	}

	return data, nil
}

// SetupI18n sets up the I18n translations for SSR.
func (s *ServerT) SetupI18n() {
	i18nBundle := i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	locales, err := ioutil.ReadDir("app/" + LocaleDir)
	if err != nil {
		f, err := s.assets.Open("/" + SSRAssetsDir + LocaleDir)
		if err != nil {
			support.Logger.Fatal(err)
		}

		locales, err = f.Readdir(-1)
	}

	for _, l := range locales {
		localePath := LocaleDir + "/" + l.Name()
		data, err := ioutil.ReadFile("app/" + localePath)
		if err != nil {
			if os.IsNotExist(err) {
				f, err := s.assets.Open("/" + SSRAssetsDir + localePath)
				if err != nil {
					support.Logger.Fatal(err)
				}

				data, err = ioutil.ReadAll(f)
				if err != nil {
					support.Logger.Fatal(err)
				}
			}
		}

		i18nBundle.MustParseMessageFileBytes(data, l.Name())
	}

	s.Use(middleware.I18n(i18nBundle))
}
