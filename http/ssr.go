package http

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	// SSRRootDebug is the root folder for debug build.
	SSRRootDebug = "app"

	// SSRRootRelease is the root folder for release build.
	SSRRootRelease = ".ssr"

	// SSRView is the views folder.
	SSRView = "views"

	// SSRLocale is the locales folder.
	SSRLocale = "locales"

	ssrRoot          = SSRRootDebug
	reservedViewDirs = []string{"layouts", "shared"}
)

func getCommonTemplates(assets http.FileSystem, build, path string) ([]string, error) {
	var (
		fis []os.FileInfo
		err error
	)

	tpls := []string{}
	if build == "debug" {
		if fis, err = ioutil.ReadDir(path); err != nil {
			return nil, err
		}
	} else {
		var file http.File
		path = "/" + path
		if file, err = assets.Open(path); err != nil {
			return nil, err
		}

		if fis, err = file.Readdir(-1); err != nil {
			return nil, err
		}
	}

	for _, fi := range fis {
		if fi.IsDir() == true {
			continue
		}

		data, err := getTemplateContent(assets, build, path+"/"+fi.Name())
		if err != nil {
			return nil, err
		}

		tpls = append(tpls, data)
	}

	return tpls, nil
}

func getTemplateContent(assets http.FileSystem, build, path string) (string, error) {
	var data []byte
	if build == "debug" {
		data, _ := ioutil.ReadFile(path)
		return string(data), nil
	}

	file, err := assets.Open(path)
	if err != nil {
		return "", err
	}

	data, _ = ioutil.ReadAll(file)
	return string(data), nil
}

// InitSSR initiates the SSR setup.
func (s *ServerT) InitSSR(vh template.FuncMap) error {
	if support.Build == "release" {
		ssrRoot = SSRRootRelease
	}

	s.ViewHelper = vh

	if err := s.initSSRLocale(); err != nil {
		return err
	}

	if err := s.initSSRView(); err != nil {
		return err
	}

	return nil
}

func (s *ServerT) initSSRView() error {
	var (
		fis []os.FileInfo
		err error
	)

	viewDir := ssrRoot + "/" + SSRView

	// We will always read from local file system when it's debug build. Otherwise, read from the bind assets.
	if support.Build == "debug" {
		if fis, err = ioutil.ReadDir(viewDir); err != nil {
			return err
		}
	} else {
		viewDir = "/" + viewDir

		var file http.File
		if file, err = s.Assets.Open(viewDir); err != nil {
			return err
		}

		fis, err = file.Readdir(-1)
	}

	commonTpls := []string{}
	for _, fi := range fis {
		// We should only see directories in `app/views`.
		if fi.IsDir() == false {
			continue
		}

		if support.Contains(reservedViewDirs, fi.Name()) == true {
			tpls, err := getCommonTemplates(s.Assets, support.Build, viewDir+"/"+fi.Name())
			if err != nil {
				return err
			}

			commonTpls = append(commonTpls, tpls...)
		}
	}

	for _, fi := range fis {
		if fi.IsDir() == false || support.Contains(reservedViewDirs, fi.Name()) == true {
			continue
		}

		var fileInfos []os.FileInfo
		targetDir := viewDir + "/" + fi.Name()
		if support.Build == "debug" {
			if fileInfos, err = ioutil.ReadDir(targetDir); err != nil {
				return err
			}
		} else {
			var file http.File
			if file, err = s.Assets.Open(targetDir); err != nil {
				return err
			}

			if fileInfos, err = file.Readdir(-1); err != nil {
				return err
			}
		}

		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() == true {
				continue
			}

			viewName := fi.Name() + "/" + fileInfo.Name()
			targetFn := targetDir + "/" + fileInfo.Name()
			data, err := getTemplateContent(s.Assets, support.Build, targetFn)
			if err != nil {
				return err
			}

			commonTplsCopy := make([]string, len(commonTpls))
			copy(commonTplsCopy, commonTpls)
			viewContent := append(commonTplsCopy, data)
			s.HTMLRenderer.AddFromStringsFuncs(viewName, s.ViewHelper, viewContent...)
		}
	}

	return nil
}

func (s *ServerT) initSSRLocale() error {
	i18nBundle := i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var (
		localeFiles []os.FileInfo
		data        []byte
		err         error
	)
	localeDir := ssrRoot + "/" + SSRLocale

	// Try getting all the locale files from `app/locales`, but fallback to `assets` http.FileSystem.
	if support.Build == "debug" {
		localeFiles, err = ioutil.ReadDir(localeDir)
		if err != nil {
			return err
		}
	} else {
		localeDir = "/" + localeDir
		file, err := s.Assets.Open(localeDir)
		if err != nil {
			return err
		}

		localeFiles, _ = file.Readdir(-1)
	}

	for _, localeFile := range localeFiles {
		localeFn := localeFile.Name()

		if support.Build == "debug" {
			data, _ = ioutil.ReadFile(localeDir + "/" + localeFn)
		} else {
			file, err := s.Assets.Open(localeDir + "/" + localeFn)
			if err != nil {
				return err
			}

			data, _ = ioutil.ReadAll(file)
		}

		i18nBundle.MustParseMessageFileBytes(data, localeFn)
	}

	s.Router.Use(middleware.I18n(i18nBundle))
	return nil
}
