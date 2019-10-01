package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type I18nSuiteT struct {
	test.SuiteT
	Recorder *httptest.ResponseRecorder
}

func (s *I18nSuiteT) SetupTest() {
	s.Recorder = httptest.NewRecorder()
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

	I18n(i18n.NewBundle(language.English))(ctx)
	s.NotNil(I18nLocalizer(ctx))
}

func TestI18n(t *testing.T) {
	test.Run(t, new(I18nSuiteT))
}
