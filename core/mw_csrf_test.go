package core

import (
	"crypto/tls"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/appist/appy/test"
	"github.com/gorilla/securecookie"
)

type CSRFSuite struct {
	test.Suite
	config   AppConfig
	logger   *AppLogger
	recorder *httptest.ResponseRecorder
}

func (s *CSRFSuite) SetupTest() {
	Build = "debug"
	s.logger, _ = newLogger(newLoggerConfig())
	s.config, _ = newConfig(nil, s.logger)
	s.config.HTTPSessionSecrets = [][]byte{[]byte("481e5d98a31585148b8b1dfb6a3c0465")}
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.recorder = httptest.NewRecorder()
	csrfSecureCookie = securecookie.New([]byte("481e5d98a31585148b8b1dfb6a3c0465"), nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(s.config.HTTPCSRFCookieMaxAge)
}

func (s *CSRFSuite) TearDownTest() {
}

func (s *CSRFSuite) TestSkipCheckContextKey() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, s.config, s.logger)
	_, exists := ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(false, exists)

	ctx, _ = test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("x-api-only", "1")
	csrfHandler(ctx, s.config, s.logger)
	_, exists = ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(true, exists)
}

func (s *CSRFSuite) TestTokenAndFieldNameContextKey() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
	}
	csrfHandler(ctx, s.config, s.logger)

	_, exists := ctx.Get(csrfCtxTokenKey)
	s.Equal(true, exists)

	_, exists = ctx.Get(csrfCtxFieldNameKey)
	s.Equal(true, exists)
	s.Equal("Cookie", ctx.Writer.Header().Get("Vary"))
}

func (s *CSRFSuite) TestRender403IfGenerateTokenError() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, errors.New("no token") }
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errors.New("no token"), ctx.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuite) TestRender403IfNoTokenGenerated() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, nil }
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfNoToken, ctx.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuite) TestRender403IfSecretKeyMissing() {
	csrfSecureCookie = securecookie.New(nil, nil)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal("securecookie: hash key is not set", ctx.Errors.Last().Err.Error())
}

func (s *CSRFSuite) TestRender403IfNoRefererOverHTTPSConn() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfNoReferer, ctx.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfMismatchRefererOverHTTPSConn() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	ctx.Request.Header.Set("Referer", "http://localhost")
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadReferer, ctx.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfNoCSRFTokenInCookie() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfTokenIsInvalid() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "test")
	ctx.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: string(realToken)})
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuite) TestTokenIsNotDecodable() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "XXXXXaGVsbG8=")
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(403, ctx.Writer.Status())
	s.Equal(errCsrfBadToken, ctx.Errors.Last().Err)
}

func (s *CSRFSuite) TestTokenIsValidInHeader() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuite) TestTokenIsValidInPostFormValue() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header:   map[string][]string{},
		Method:   "POST",
		PostForm: url.Values{},
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.PostForm.Add(csrfTemplateFieldName(ctx), authenticityToken)
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuite) TestTokenIsValidInMultipartForm() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		MultipartForm: &multipart.Form{
			Value: map[string][]string{},
		},
	}
	ctx.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	ctx.Request.MultipartForm.Value[csrfTemplateFieldName(ctx)] = []string{authenticityToken}
	csrfHandler(ctx, s.config, s.logger)

	s.Equal(200, ctx.Writer.Status())
}

func (s *CSRFSuite) TestCSRFSkipCheck() {
	ctx, _ := test.CreateTestContext(s.recorder)
	_, exists := ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(false, exists)

	ctx, _ = test.CreateTestContext(s.recorder)
	CSRFSkipCheck()(ctx)
	_, exists = ctx.Get(csrfCtxSkipCheckKey)
	s.Equal(true, exists)
}

func (s *CSRFSuite) TestCSRFTemplateField() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal(`<input type="hidden" name="authenticity_token" value="test">`, CSRFTemplateField(ctx))

	ctx, _ = test.CreateTestContext(s.recorder)
	ctx.Set(csrfCtxTokenKey, "")
	s.Equal(`<input type="hidden" name="authenticity_token" value="">`, CSRFTemplateField(ctx))

	ctx, _ = test.CreateTestContext(s.recorder)
	ctx.Set(csrfCtxFieldNameKey, "my_authenticity_token")
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal(`<input type="hidden" name="my_authenticity_token" value="test">`, CSRFTemplateField(ctx))
}

func (s *CSRFSuite) TestCSRFToken() {
	ctx, _ := test.CreateTestContext(s.recorder)
	s.Equal("", CSRFToken(ctx))

	ctx, _ = test.CreateTestContext(s.recorder)
	ctx.Set(csrfCtxTokenKey, "test")
	s.Equal("test", CSRFToken(ctx))
}

func (s *CSRFSuite) TestCSRFMiddleware() {
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	CSRF(s.config, s.logger)(ctx)
	s.NotNil(csrfSecureCookie)
}

func TestCSRF(t *testing.T) {
	test.Run(t, new(CSRFSuite))
}
