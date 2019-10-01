package middleware

import (
	"crypto/tls"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	csrfSecureCookie = securecookie.New([]byte("481e5d98a31585148b8b1dfb6a3c0465"), nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(s.Config.HTTPCSRFCookieMaxAge)
}

func (s *CSRFSuiteT) TestSkipCheckContextKey() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, s.Config)
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

func (s *CSRFSuiteT) TestTokenAndFieldNameContextKey() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
	}
	csrfHandler(ctx, support.Config)
	_, exists := ctx.Get(csrfCtxTokenKey)
	s.Equal(true, exists)
	_, exists = ctx.Get(csrfCtxFieldNameKey)
	s.Equal(true, exists)
	s.Equal("Cookie", ctx.Writer.Header().Get("Vary"))
}

func (s *CSRFSuiteT) TestRender403IfGenerateTokenError() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, errors.New("no token") }
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errors.New("no token"), ctx.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuiteT) TestRender403IfNoTokenGenerated() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, nil }
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfNoToken, ctx.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuiteT) TestRender403IfSecretKeyMissing() {
	csrfSecureCookie = securecookie.New(nil, nil)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}

	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal("securecookie: hash key is not set", ctx.Errors.Last().Err.Error())
}

func (s *CSRFSuiteT) TestRender403IfNoRefererOverHTTPSConn() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfNoReferer, ctx.Errors.Last().Err)
}

func (s *CSRFSuiteT) TestRender403IfMismatchRefererOverHTTPSConn() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	ctx.Request.Header.Set("Referer", "http://0.0.0.0")
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadReferer, ctx.Errors.Last().Err)
}

func (s *CSRFSuiteT) TestRender403IfNoCSRFTokenInCookie() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.Header.Set(s.Config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuiteT) TestRender403IfTokenIsInvalid() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.Header.Set(s.Config.HTTPCSRFRequestHeader, "test")
	ctx.Request.AddCookie(&http.Cookie{Name: s.Config.HTTPCSRFCookieName, Value: string(realToken)})
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuiteT) TestTokenIsNotDecodable() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.Config.HTTPCSRFCookieName, realToken)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.Config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.Header.Set(s.Config.HTTPCSRFRequestHeader, "XXXXXaGVsbG8=")
	csrfHandler(ctx, s.Config)
	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuiteT) TestTokenIsValidInHeader() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.Config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.Config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.Header.Set(s.Config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(ctx, s.Config)
	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuiteT) TestTokenIsValidInPostFormValue() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.Config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header:   map[string][]string{},
		Method:   "POST",
		PostForm: url.Values{},
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.Config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.PostForm.Add(csrfTemplateFieldName(ctx), authenticityToken)
	csrfHandler(ctx, s.Config)
	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuiteT) TestTokenIsValidInMultipartForm() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.Config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		MultipartForm: &multipart.Form{
			Value: map[string][]string{},
		},
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.Config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.MultipartForm.Value[csrfTemplateFieldName(ctx)] = []string{authenticityToken}
	csrfHandler(ctx, s.Config)
	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuiteT) TestCSRFSkipCheck() {
	ctx, _ := test.CreateContext(s.Recorder)
	_, exists := ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(false, exists)

	ctx, _ = test.CreateContext(s.Recorder)
	CSRFSkipCheck()(ctx)
	_, exists = ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(true, exists)
}

func (s *CSRFSuiteT) TestCSRFTemplateField() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal(`<input type="hidden" name="authenticity_token" value="test">`, CSRFTemplateField(ctx))

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Set(csrfCtxTokenKey, "")
	s.Equal(`<input type="hidden" name="authenticity_token" value="">`, CSRFTemplateField(ctx))

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Set(csrfCtxFieldNameKey, "my_authenticity_token")
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal(`<input type="hidden" name="my_authenticity_token" value="test">`, CSRFTemplateField(ctx))
}

func (s *CSRFSuiteT) TestCSRFToken() {
	ctx, _ := test.CreateContext(s.Recorder)
	s.Equal("", CSRFToken(ctx))

	ctx, _ = test.CreateContext(s.Recorder)
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal("test", CSRFToken(ctx))
}

func (s *CSRFSuiteT) TestCSRFMiddleware() {
	s.Config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	CSRF(s.Config)(ctx)
	s.NotNil(csrfSecureCookie)
}

func TestCSRF(t *testing.T) {
	test.Run(t, new(CSRFSuiteT))
}
