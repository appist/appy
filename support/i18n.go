package support

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type (
	// I18n manages the translations.
	I18n struct {
		bundle *i18n.Bundle
		config *Config
		logger *Logger
	}
)

// NewI18n initializes the I18n instance.
func NewI18n(assets *Assets, config *Config, logger *Logger) *I18n {
	languageTag := language.MustParse("en")
	if config != nil && config.I18nDefaultLocale != "" {
		languageTag = language.MustParse(config.I18nDefaultLocale)
	}

	bundle := i18n.NewBundle(languageTag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	fis, err := assets.ReadDir(assets.Layout()["locale"])
	if err != nil {
		panic(err)
	}

	for _, fi := range fis {
		filename := assets.Layout()["locale"] + "/" + fi.Name()
		data, err := assets.ReadFile(filename, true)

		if err != nil {
			panic(err)
		}

		bundle.MustParseMessageFileBytes(data, fi.Name())
	}

	return &I18n{
		bundle: bundle,
		config: config,
		logger: logger,
	}
}

// Bundle returns the I18n bundle which contains the loaded locales.
func (i *I18n) Bundle() *i18n.Bundle {
	return i.bundle
}

// Locales returns all the available locales.
func (i *I18n) Locales() []string {
	locales := []string{}

	for _, tag := range i.bundle.LanguageTags() {
		locales = append(locales, tag.String())
	}

	return locales
}

// T translates a message based on the given key and locale.
func (i *I18n) T(key string, args ...interface{}) string {
	var data H
	count := -1
	locale := i.config.I18nDefaultLocale

	for _, arg := range args {
		switch v := arg.(type) {
		case H:
			data = v
		case int:
			count = v
			switch count {
			case 0:
				key = key + ".Zero"
			case 1:
				key = key + ".One"
			default:
				key = key + ".Other"
			}
		case string:
			locale = v
		}
	}

	if count > -1 {
		data["Count"] = count
	}

	localizer := i18n.NewLocalizer(i.bundle, locale)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: key, TemplateData: data})
	if err != nil {
		i.logger.Warn(err)
		return ""
	}

	return msg
}
