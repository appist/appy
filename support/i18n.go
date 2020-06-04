package support

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// I18n manages the application translations.
type I18n struct {
	bundle *i18n.Bundle
	config *Config
	logger *Logger
}

const validateErrorPrefix = "errors.messages."

// NewI18n initializes the I18n instance.
func NewI18n(asset *Asset, config *Config, logger *Logger) *I18n {
	languageTag := language.MustParse("en")
	if config != nil && config.I18nDefaultLocale != "" {
		languageTag = language.MustParse(config.I18nDefaultLocale)
	}

	bundle := i18n.NewBundle(languageTag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	fis, err := asset.ReadDir(asset.Layout().Locale())
	if err != nil {
		panic(err)
	}

	for _, fi := range fis {
		filename := asset.Layout().Locale() + "/" + fi.Name()
		data, _ := asset.ReadFile(filename)
		_, _ = bundle.ParseMessageFileBytes(data, fi.Name())
	}

	addDefaultValidationErrors(bundle)

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
	data := H{}
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

func addDefaultValidationErrors(bundle *i18n.Bundle) {
	localizer := i18n.NewLocalizer(bundle, "en")
	messages := map[string]*i18n.Message{
		validateErrorPrefix + "eq": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} the value(number/string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "gt": {
			Other: "{{.Field}} must be greater than {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "gte": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "len": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "lt": {
			Other: "{{.Field}} must be less than {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "lte": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "max": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "min": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} the value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "ne": {
			Other: "{{.Field}} must not be equal to {{.ExpectedValue}} the value(number/string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "required": {
			Other: "{{.Field}} must not be blank",
		},
	}

	for id, message := range messages {
		_, err := localizer.LocalizeMessage(&i18n.Message{ID: id})

		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			message.ID = id
			bundle.AddMessages(language.English, message)
		}
	}
}
