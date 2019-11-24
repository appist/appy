package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/internal/test"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type I18nSuite struct {
	test.Suite
	i18nBundle *i18n.Bundle
	recorder   *httptest.ResponseRecorder
}

func (s *I18nSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
	s.i18nBundle = i18n.NewBundle(language.English)
	s.i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.i18nBundle.MustLoadMessageFile("testdata/locales/en.yml")
	s.i18nBundle.MustLoadMessageFile("testdata/locales/zh-CN.yml")
	s.i18nBundle.MustLoadMessageFile("testdata/locales/zh-TW.yml")
}

func (s *I18nSuite) TearDownTest() {
}

func (s *I18nSuite) TestI18nCtxKeyIsNotSetIfNotConfigured() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	_, exists := ctx.Get(i18nCtxKey.String())
	s.Equal(false, exists)
}

func (s *I18nSuite) TestI18nCtxKeyIsSetIfConfigured() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(i18n.NewBundle(language.English))(ctx)
	_, exists := ctx.Get(i18nCtxKey.String())
	s.Equal(true, exists)
}

func (s *I18nSuite) TestI18nLocalizerIsNilIfNotConfigured() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	s.Nil(I18nLocalizer(ctx))
}

func (s *I18nSuite) TestI18nLocalizerIsNotNilIfConfigured() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(s.i18nBundle)(ctx)
	s.NotNil(I18nLocalizer(ctx))
}

func (s *I18nSuite) TestI18nLocaleIsENByDefault() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(s.i18nBundle)(ctx)
	s.Equal("en", I18nLocale(ctx))
}

func (s *I18nSuite) TestI18nLocaleIsSetByAcceptLanguageHeader() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-TW")

	I18n(s.i18nBundle)(ctx)
	s.Equal("zh-TW", I18nLocale(ctx))
}

func (s *I18nSuite) TestT() {
	ctx, _ := test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	I18n(s.i18nBundle)(ctx)

	s.Equal("", T(ctx, "missing", -1, nil))
	s.Equal("", T(ctx, "missing", -1, nil))
	s.Equal("Password", T(ctx, "password", -1, nil))
	s.Equal("John Doe has no message.", T(ctx, "message", 0, H{"Name": "John Doe"}))
	s.Equal("John Doe has 1 message.", T(ctx, "message", 1, H{"Name": "John Doe"}))
	s.Equal("John Doe has 2 messages.", T(ctx, "message", 2, H{"Name": "John Doe"}))

	ctx, _ = test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-CN")
	I18n(s.i18nBundle)(ctx)

	s.Equal("密码", T(ctx, "password", -1, nil))
	s.Equal("John Doe没有讯息。", T(ctx, "message", 0, H{"Name": "John Doe"}))
	s.Equal("John Doe有1则讯息。", T(ctx, "message", 1, H{"Name": "John Doe"}))
	s.Equal("John Doe有2则讯息。", T(ctx, "message", 2, H{"Name": "John Doe"}))

	ctx, _ = test.CreateHTTPContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-TW")
	I18n(s.i18nBundle)(ctx)

	s.Equal("密碼", T(ctx, "password", -1, nil))
	s.Equal("John Doe沒有訊息。", T(ctx, "message", 0, H{"Name": "John Doe"}))
	s.Equal("John Doe有1則訊息。", T(ctx, "message", 1, H{"Name": "John Doe"}))
	s.Equal("John Doe有2則訊息。", T(ctx, "message", 2, H{"Name": "John Doe"}))
}

func TestI18nSuite(t *testing.T) {
	test.RunSuite(t, new(I18nSuite))
}
