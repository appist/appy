package appy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/securecookie"
)

type CSRFSuite struct {
	TestSuite
	asset    *Asset
	config   *Config
	logger   *Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
	support  Supporter
}

func (s *CSRFSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &Support{}
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.asset = NewAsset(http.Dir("testdata"), nil)
	s.config = NewConfig(s.asset, s.logger, s.support)
	s.recorder = httptest.NewRecorder()
	csrfSecureCookie = securecookie.New([]byte("481e5d98a31585148b8b1dfb6a3c0465"), nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(s.config.HTTPCSRFCookieMaxAge)
}

func (s *CSRFSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *CSRFSuite) TestSkipCheckContextKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(c, s.config, s.logger, s.support)
	_, exists := c.Get(csrfSkipCheckCtxKey.String())
	s.Equal(false, exists)

	c, _ = NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("x-api-only", "1")
	csrfHandler(c, s.config, s.logger, s.support)
	_, exists = c.Get(csrfSkipCheckCtxKey.String())
	s.Equal(true, exists)
}

func (s *CSRFSuite) TestTokenAndFieldNameContextKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
	}
	csrfHandler(c, s.config, s.logger, s.support)

	_, exists := c.Get(csrfTokenCtxKey.String())
	s.Equal(true, exists)

	_, exists = c.Get(csrfFieldNameCtxKey.String())
	s.Equal(true, exists)
	s.Equal("Cookie", c.Writer.Header().Get("Vary"))
}

func (s *CSRFSuite) TestRender403IfGenerateTokenError() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, errors.New("no token") }
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errors.New("no token"), c.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuite) TestRender403IfNoTokenGenerated() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, nil }
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfNoToken, c.Errors.Last().Err)
	generateRandomBytes = oldGRB
}

func (s *CSRFSuite) TestRender403IfSecretKeyMissing() {
	csrfSecureCookie = securecookie.New(nil, nil)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal("securecookie: hash key is not set", c.Errors.Last().Err.Error())
}

func (s *CSRFSuite) TestRender403IfNoRefererOverHTTPSConn() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfNoReferer, c.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfMismatchRefererOverHTTPSConn() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	c.Request.Header.Set("Referer", "http://localhost")
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfBadReferer, c.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfNoCSRFTokenInCookie() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	authenticityToken := getCSRFMaskedToken(realToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfBadToken, c.Errors.Last().Err)
}

func (s *CSRFSuite) TestRender403IfTokenIsInvalid() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "test")
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: string(realToken)})
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfBadToken, c.Errors.Last().Err)
}

func (s *CSRFSuite) TestTokenIsNotDecodable() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "XXXXXaGVsbG8=")
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCsrfBadToken, c.Errors.Last().Err)
}

func (s *CSRFSuite) TestTokenIsValidInHeader() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *CSRFSuite) TestTokenIsValidInPostFormValue() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header:   map[string][]string{},
		Method:   "POST",
		PostForm: url.Values{},
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	c.Request.PostForm.Add(csrfTemplateFieldName(c), authenticityToken)
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *CSRFSuite) TestTokenIsValidInMultipartForm() {
	realToken, _ := generateRandomBytes(csrfTokenLength)
	encRealToken, _ := csrfSecureCookie.Encode(s.config.HTTPCSRFCookieName, realToken)
	authenticityToken := getCSRFMaskedToken(realToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		MultipartForm: &multipart.Form{
			Value: map[string][]string{},
		},
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encRealToken})
	c.Request.MultipartForm.Value[csrfTemplateFieldName(c)] = []string{authenticityToken}
	csrfHandler(c, s.config, s.logger, s.support)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *CSRFSuite) TestCSRFSkipCheck() {
	c, _ := NewTestContext(s.recorder)
	_, exists := c.Get(csrfSkipCheckCtxKey.String())
	s.Equal(false, exists)

	c, _ = NewTestContext(s.recorder)
	CSRFSkipCheck()(c)
	_, exists = c.Get(csrfSkipCheckCtxKey.String())
	s.Equal(true, exists)
}

func (s *CSRFSuite) TestCSRFTemplateField() {
	c, _ := NewTestContext(s.recorder)
	c.Set(csrfTokenCtxKey.String(), "test")
	s.Equal(`<input type="hidden" name="authenticity_token" value="test">`, c.CSRFTemplateField())

	c, _ = NewTestContext(s.recorder)
	c.Set(csrfTokenCtxKey.String(), "")
	s.Equal(`<input type="hidden" name="authenticity_token" value="">`, c.CSRFTemplateField())

	c, _ = NewTestContext(s.recorder)
	c.Set(csrfFieldNameCtxKey.String(), "my_authenticity_token")
	c.Set(csrfTokenCtxKey.String(), "test")
	s.Equal(`<input type="hidden" name="my_authenticity_token" value="test">`, c.CSRFTemplateField())
}

func (s *CSRFSuite) TestCSRFToken() {
	c, _ := NewTestContext(s.recorder)
	s.Equal("", c.CSRFToken())

	c, _ = NewTestContext(s.recorder)
	c.Set(csrfTokenCtxKey.String(), "test")
	s.Equal("test", c.CSRFToken())
}

func (s *CSRFSuite) TestCSRFMiddleware() {
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	CSRF(s.config, s.logger, s.support)(c)
	s.NotNil(csrfSecureCookie)
}

func TestCSRF(t *testing.T) {
	RunTestSuite(t, new(CSRFSuite))
}
