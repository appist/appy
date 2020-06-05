package support

import (
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
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
		bundle.MustParseMessageFileBytes(data, fi.Name())
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

func (i *I18n) GetValidationErrors(err error, locale string) []error {
	errs := []error{}
	verrs := err.(validator.ValidationErrors)

	for _, verr := range verrs {
		var (
			field, message                                                                string
			fieldKeyBuilder, generalKeyBuilder, modelAttributeKeyBuilder, paramKeyBuilder strings.Builder
		)

		args := []interface{}{}
		if locale != "" {
			args = append(args, locale)
		}

		fieldKeyBuilder.WriteString("models.")
		fieldKeyBuilder.WriteString(verr.StructNamespace())

		field = i.T(fieldKeyBuilder.String(), args...)
		if field == "" {
			field = verr.StructNamespace()
		}

		param := verr.Param()
		if ArrayContains([]string{"eqfield"}, verr.Tag()) {
			ns := strings.Split(verr.StructNamespace(), ".")

			if len(ns) > 1 {
				paramKeyBuilder.WriteString("models.")
				paramKeyBuilder.WriteString(ns[0])
				paramKeyBuilder.WriteString(".")
				paramKeyBuilder.WriteString(param)
				transParam := i.T("models."+ns[0]+"."+param, args...)

				if transParam != "" {
					param = transParam
				}
			}
		}

		args = append(args, H{
			"ExactValue":    verr.Value(),
			"ExpectedValue": param,
			"Field":         field,
		})

		modelAttributeKeyBuilder.WriteString("errors.models.")
		modelAttributeKeyBuilder.WriteString(verr.StructNamespace())
		modelAttributeKeyBuilder.WriteString(".")
		modelAttributeKeyBuilder.WriteString(verr.Tag())

		message = i.T(modelAttributeKeyBuilder.String(), args...)
		if message == "" {
			generalKeyBuilder.WriteString("errors.messages.")
			generalKeyBuilder.WriteString(verr.Tag())
			message = i.T(generalKeyBuilder.String(), args...)
		}

		errs = append(errs, errors.New(message))
	}

	return errs
}

func addDefaultValidationErrors(bundle *i18n.Bundle) {
	localizer := i18n.NewLocalizer(bundle, "en")
	messages := map[string]*i18n.Message{
		validateErrorPrefix + "alpha": {
			Other: "{{.Field}} must contain ASCII alpha characters only",
		},
		validateErrorPrefix + "alphanum": {
			Other: "{{.Field}} must contain ASCII alphanumeric characters only",
		},
		validateErrorPrefix + "alphaunicode": {
			Other: "{{.Field}} must contain ASCII unicode alpha characters only",
		},
		validateErrorPrefix + "alphanumunicode": {
			Other: "{{.Field}} must contain ASCII unicode alphanumeric characters only",
		},
		validateErrorPrefix + "ascii": {
			Other: "{{.Field}} must contain ASCII characters only",
		},
		validateErrorPrefix + "base64": {
			Other: "{{.Field}} must be a valid base64 string",
		},
		validateErrorPrefix + "base64url": {
			Other: "{{.Field}} must be a valid base64 safe value URL string according to RFC4648 spec",
		},
		validateErrorPrefix + "btc_addr": {
			Other: "{{.Field}} must be a valid bitcoin address",
		},
		validateErrorPrefix + "btc_addr_bech32": {
			Other: "{{.Field}} must be a valid bitcoin Bech32 address",
		},
		validateErrorPrefix + "cidr": {
			Other: "{{.Field}} must be a valid CIDR address",
		},
		validateErrorPrefix + "cidrv4": {
			Other: "{{.Field}} must be a valid v4 CIDR address",
		},
		validateErrorPrefix + "cidrv6": {
			Other: "{{.Field}} must be a valid v6 CIDR address",
		},
		validateErrorPrefix + "contains": {
			Other: "{{.Field}} must contain '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "containsfield": {
			Other: "{{.Field}} must contain {{.ExpectedValue}} field's value",
		},
		validateErrorPrefix + "containsany": {
			Other: "{{.Field}} must contain 1 of the Unicode code points in '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "containsrune": {
			Other: "{{.Field}} must contain '{{.ExpectedValue}}' rune",
		},
		validateErrorPrefix + "datauri": {
			Other: "{{.Field}} must be a valid DataURI",
		},
		validateErrorPrefix + "dir": {
			Other: "{{.Field}} must be a valid directory path that exists on the machine",
		},
		validateErrorPrefix + "email": {
			Other: "{{.Field}} must be a valid email",
		},
		validateErrorPrefix + "endswith": {
			Other: "{{.Field}} must end with '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "eq": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} in value(number/string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "eqfield": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "eqcsfield": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "eth_addr": {
			Other: "{{.Field}} must be a valid ethereum address",
		},
		validateErrorPrefix + "excludes": {
			Other: "{{.Field}} must not contain '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "excludesfield": {
			Other: "{{.Field}} must not contain {{.ExpectedValue}} field's value",
		},
		validateErrorPrefix + "excludesall": {
			Other: "{{.Field}} must not contain any of the Unicode code points in '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "excludesrune": {
			Other: "{{.Field}} must not contain '{{.ExpectedValue}}' rune",
		},
		validateErrorPrefix + "file": {
			Other: "{{.Field}} must be a valid file path that exists on the machine",
		},
		validateErrorPrefix + "fqdn": {
			Other: "{{.Field}} must be a valid FQDN",
		},
		validateErrorPrefix + "gt": {
			Other: "{{.Field}} must be greater than {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "gtfield": {
			Other: "{{.Field}} must be greater than {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "gtcsfield": {
			Other: "{{.Field}} must be greater than {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "gte": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "gtefield": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "gtecsfield": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "hexadecimal": {
			Other: "{{.Field}} must be a valid hexadecimal",
		},
		validateErrorPrefix + "hexcolor": {
			Other: "{{.Field}} must be a valid hex color with a # prefix",
		},
		validateErrorPrefix + "hostname": {
			Other: "{{.Field}} must be a valid hostname according to RFC 952",
		},
		validateErrorPrefix + "hostname_rfc1123": {
			Other: "{{.Field}} must be a valid hostname according to RFC 1123",
		},
		validateErrorPrefix + "hsl": {
			Other: "{{.Field}} must be a valid HSL color",
		},
		validateErrorPrefix + "hsla": {
			Other: "{{.Field}} must be a valid HSLA color",
		},
		validateErrorPrefix + "html": {
			Other: "{{.Field}} must be a valid HTML element tag",
		},
		validateErrorPrefix + "html_encoded": {
			Other: "{{.Field}} must be a valid HTML element tag in decimal/hexadecimal format",
		},
		validateErrorPrefix + "ip": {
			Other: "{{.Field}} must be a valid IP address",
		},
		validateErrorPrefix + "ipv4": {
			Other: "{{.Field}} must be a valid v4 IP address",
		},
		validateErrorPrefix + "ipv6": {
			Other: "{{.Field}} must be a valid v6 IP address",
		},
		validateErrorPrefix + "ip_addr": {
			Other: "{{.Field}} must be a valid resolvable IP address",
		},
		validateErrorPrefix + "ip4_addr": {
			Other: "{{.Field}} must be a valid resolvable v4 IP address",
		},
		validateErrorPrefix + "ip6_addr": {
			Other: "{{.Field}} must be a valid resolvable v6 IP address",
		},
		validateErrorPrefix + "isbn": {
			Other: "{{.Field}} must be a valid ISBN10 or ISBN13 value",
		},
		validateErrorPrefix + "isbn10": {
			Other: "{{.Field}} must be a valid ISBN10 value",
		},
		validateErrorPrefix + "isbn13": {
			Other: "{{.Field}} must be a valid ISBN13 value",
		},
		validateErrorPrefix + "latitude": {
			Other: "{{.Field}} must be a valid latitude",
		},
		validateErrorPrefix + "len": {
			Other: "{{.Field}} must be equal to {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "longitude": {
			Other: "{{.Field}} must be a valid longitude",
		},
		validateErrorPrefix + "lt": {
			Other: "{{.Field}} must be less than {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "ltfield": {
			Other: "{{.Field}} must be less than {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "ltcsfield": {
			Other: "{{.Field}} must be less than {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "lte": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "ltefield": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "ltecsfield": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "mac": {
			Other: "{{.Field}} must be a valid MAC address",
		},
		validateErrorPrefix + "max": {
			Other: "{{.Field}} must be less than or equal to {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "min": {
			Other: "{{.Field}} must be greater than or equal to {{.ExpectedValue}} in value(number), length(string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "multibyte": {
			Other: "{{.Field}} must contain 1 or more multi-byte characters",
		},
		validateErrorPrefix + "ne": {
			Other: "{{.Field}} must not be equal to {{.ExpectedValue}} in value(number/string) or number of items(arrays/slices/maps)",
		},
		validateErrorPrefix + "nefield": {
			Other: "{{.Field}} must not be equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "necsfield": {
			Other: "{{.Field}} must not be equal to {{.ExpectedValue}} field",
		},
		validateErrorPrefix + "numeric": {
			Other: "{{.Field}} must be a valid numeric",
		},
		validateErrorPrefix + "oneof": {
			Other: "{{.Field}} must be one of the values in [{{.ExpectedValue}}]",
		},
		validateErrorPrefix + "printascii": {
			Other: "{{.Field}} must contain printable ASCII characters only",
		},
		validateErrorPrefix + "required": {
			Other: "{{.Field}} must not be blank",
		},
		validateErrorPrefix + "startswith": {
			Other: "{{.Field}} must start with '{{.ExpectedValue}}'",
		},
		validateErrorPrefix + "rgb": {
			Other: "{{.Field}} must be a valid RGB color",
		},
		validateErrorPrefix + "rgba": {
			Other: "{{.Field}} must be a valid RGBA color",
		},
		validateErrorPrefix + "ssn": {
			Other: "{{.Field}} must be a valid U.S. Social Security Number",
		},
		validateErrorPrefix + "tcp_addr": {
			Other: "{{.Field}} must be a valid resolvable TCP address",
		},
		validateErrorPrefix + "tcp4_addr": {
			Other: "{{.Field}} must be a valid resolvable v4 TCP address",
		},
		validateErrorPrefix + "tcp6_addr": {
			Other: "{{.Field}} must be a valid resolvable v6 TCP address",
		},
		validateErrorPrefix + "udp_addr": {
			Other: "{{.Field}} must be a valid resolvable UDP address",
		},
		validateErrorPrefix + "udp4_addr": {
			Other: "{{.Field}} must be a valid resolvable v4 UDP address",
		},
		validateErrorPrefix + "udp6_addr": {
			Other: "{{.Field}} must be a valid resolvable v6 UDP address",
		},
		validateErrorPrefix + "unique": {
			Other: "the values in {{.Field}} must be unique",
		},
		validateErrorPrefix + "unix_addr": {
			Other: "{{.Field}} must be a valid Unix address",
		},
		validateErrorPrefix + "uri": {
			Other: "{{.Field}} must be a valid URI",
		},
		validateErrorPrefix + "url": {
			Other: "{{.Field}} must be a valid URL",
		},
		validateErrorPrefix + "url_encoded": {
			Other: "{{.Field}} must be a valid percent-encoded URL",
		},
		validateErrorPrefix + "urn_rfc2141": {
			Other: "{{.Field}} must be a valid URN according to the RFC 2141 spec",
		},
		validateErrorPrefix + "uuid": {
			Other: "{{.Field}} must be a valid UUID",
		},
		validateErrorPrefix + "uuid3": {
			Other: "{{.Field}} must be a valid v3 UUID",
		},
		validateErrorPrefix + "uuid4": {
			Other: "{{.Field}} must be a valid v4 UUID",
		},
		validateErrorPrefix + "uuid5": {
			Other: "{{.Field}} must be a valid v5 UUID",
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
