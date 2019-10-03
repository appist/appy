package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type I18nSuiteT struct {
	test.SuiteT
	I18nBundle *i18n.Bundle
	Recorder   *httptest.ResponseRecorder
}

func (s *I18nSuiteT) SetupTest() {
	s.Recorder = httptest.NewRecorder()
	s.I18nBundle = i18n.NewBundle(language.English)
	s.I18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.I18nBundle.MustLoadMessageFile("../testdata/locales/en.yml")
	s.I18nBundle.MustLoadMessageFile("../testdata/locales/zh-CN.yml")
	s.I18nBundle.MustLoadMessageFile("../testdata/locales/zh-TW.yml")
}

func (s *I18nSuiteT) TestI18nCtxKeyIsNotSetIfNotConfigured() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	_, exists := ctx.Get(I18nCtxKey)
	s.Equal(false, exists)
}

func (s *I18nSuiteT) TestI18nCtxKeyIsSetIfConfigured() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(i18n.NewBundle(language.English))(ctx)
	_, exists := ctx.Get(I18nCtxKey)
	s.Equal(true, exists)
}

func (s *I18nSuiteT) TestI18nLocalizerIsNilIfNotConfigured() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	s.Nil(I18nLocalizer(ctx))
}

func (s *I18nSuiteT) TestI18nLocalizerIsNotNilIfConfigured() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(s.I18nBundle)(ctx)
	s.NotNil(I18nLocalizer(ctx))
}

func (s *I18nSuiteT) TestI18nLocaleIsENByDefault() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	I18n(s.I18nBundle)(ctx)
	s.Equal("en", I18nLocale(ctx))
}

func (s *I18nSuiteT) TestI18nLocaleIsSetByAcceptLanguageHeader() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-TW")

	I18n(s.I18nBundle)(ctx)
	s.Equal("zh-TW", I18nLocale(ctx))
}

func (s *I18nSuiteT) TestT() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	I18n(s.I18nBundle)(ctx)

	s.Equal("", T(ctx, "missing"))
	s.Equal("", T(ctx, "missing", gin.H{"Count": 0}))
	s.Equal("Password", T(ctx, "password"))
	s.Equal("John Doe has no message.", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 0}))
	s.Equal("John Doe has 1 message.", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 1}))
	s.Equal("John Doe has 2 messages.", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 2}))

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-CN")
	I18n(s.I18nBundle)(ctx)

	s.Equal("密码", T(ctx, "password"))
	s.Equal("John Doe没有讯息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 0}))
	s.Equal("John Doe有1则讯息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 1}))
	s.Equal("John Doe有2则讯息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 2}))

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add(acceptLanguage, "zh-TW")
	I18n(s.I18nBundle)(ctx)

	s.Equal("密碼", T(ctx, "password"))
	s.Equal("John Doe沒有訊息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 0}))
	s.Equal("John Doe有1則訊息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 1}))
	s.Equal("John Doe有2則訊息。", T(ctx, "message", gin.H{"Name": "John Doe", "Count": 2}))
}

func TestI18n(t *testing.T) {
	test.Run(t, new(I18nSuiteT))
}
