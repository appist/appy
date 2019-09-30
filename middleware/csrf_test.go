package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/gorilla/securecookie"
)

type CSRFSuiteT struct {
	test.SuiteT
	Config   *support.ConfigT
	Recorder *httptest.ResponseRecorder
}

func (s *CSRFSuiteT) SetupTest() {
	s.Config = support.Config
	s.Recorder = httptest.NewRecorder()
}

func (s *CSRFSuiteT) TestSkipCheckContextKey() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	s.Panics(func() { csrfHandler(ctx, s.Config) })
	_, exists := ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(false, exists)

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("x-api-only", "1")
	csrfHandler(ctx, s.Config)
	_, exists = ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(true, exists)
}

func (s *CSRFSuiteT) TestRender403IfError() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, support.Config)
}

func (s *CSRFSuiteT) TestTokenAndFieldNameSet() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
	}
	oCookie := csrfSecureCookie
	csrfSecureCookie = securecookie.New([]byte("481e5d98a31585148b8b1dfb6a3c0465"), nil)
	csrfHandler(ctx, support.Config)
	csrfSecureCookie = oCookie
	_, exists := ctx.Get(csrfCtxTokenKey)
	s.Equal(true, exists)
	_, exists = ctx.Get(csrfCtxFieldNameKey)
	s.Equal(true, exists)
	s.Equal("Cookie", ctx.Writer.Header().Get("Vary"))
}

func TestCSRF(t *testing.T) {
	test.Run(t, new(CSRFSuiteT))
}
