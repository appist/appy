package pack

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

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/gorilla/securecookie"
)

type mdwCSRFSuite struct {
	test.Suite
	asset    *support.Asset
	config   *support.Config
	logger   *support.Logger
	buffer   *bytes.Buffer
	writer   *bufio.Writer
	recorder *httptest.ResponseRecorder
}

func (s *mdwCSRFSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.recorder = httptest.NewRecorder()
	mdwCSRFSecureCookie = securecookie.New([]byte("481e5d98a31585148b8b1dfb6a3c0465"), nil)
	mdwCSRFSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	mdwCSRFSecureCookie.MaxAge(s.config.HTTPCSRFCookieMaxAge)
}

func (s *mdwCSRFSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwCSRFSuite) TestSkipCheckContextKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	mdwCSRFHandler(c, s.config, s.logger)
	_, exists := c.Get(mdwCSRFSkipCheckCtxKey.String())
	s.Equal(false, exists)

	c, _ = NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("x-api-only", "1")
	mdwCSRFHandler(c, s.config, s.logger)
	_, exists = c.Get(mdwCSRFSkipCheckCtxKey.String())
	s.Equal(true, exists)
}

func (s *mdwCSRFSuite) TestTokenAndFieldNameContextKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "GET",
	}
	mdwCSRFHandler(c, s.config, s.logger)

	_, exists := c.Get(mdwCSRFAuthenticityTokenCtxKey.String())
	s.Equal(true, exists)

	_, exists = c.Get(mdwCSRFAuthenticityFieldNameCtxKey.String())
	s.Equal(true, exists)
	s.Equal("Cookie", c.Writer.Header().Get("Vary"))
}

func (s *mdwCSRFSuite) TestRender403IfGenerateCSRFTokenError() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, errors.New("no token") }
	defer func() { generateRandomBytes = oldGRB }()

	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errors.New("no token"), c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender403IfNoCSRFTokenGenerated() {
	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, nil }
	defer func() { generateRandomBytes = oldGRB }()

	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFNoToken, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender500IfGenerateAuthenticityTokenError() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	encCSRFToken, _ := mdwCSRFSecureCookie.Encode(s.config.HTTPCSRFCookieName, csrfToken)

	oldGRB := generateRandomBytes
	generateRandomBytes = func(n int) ([]byte, error) { return nil, errors.New("no token") }
	defer func() { generateRandomBytes = oldGRB }()

	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encCSRFToken})
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusInternalServerError, c.Writer.Status())
	s.Equal(errors.New("no token"), c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender403IfSecretKeyMissing() {
	mdwCSRFSecureCookie = securecookie.New(nil, nil)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal("securecookie: hash key is not set", c.Errors.Last().Err.Error())
}

func (s *mdwCSRFSuite) TestRender403IfNoRefererOverHTTPSConn() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFNoReferer, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender403IfMismatchRefererOverHTTPSConn() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		TLS:    &tls.ConnectionState{},
	}
	c.Request.Header.Set("Referer", "http://localhost")
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFBadReferer, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender403IfNoCSRFTokenInCookie() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	authenticityToken, _ := generateAuthenticityToken(csrfToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFBadToken, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestRender403IfTokenIsInvalid() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "test")
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: string(csrfToken)})
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFBadToken, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestAuthenticityTokenIsNotDecodable() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	encCSRFToken, _ := mdwCSRFSecureCookie.Encode(s.config.HTTPCSRFCookieName, csrfToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encCSRFToken})
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, "XXXXXaGVsbG8=")
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusForbidden, c.Writer.Status())
	s.Equal(errCSRFBadToken, c.Errors.Last().Err)
}

func (s *mdwCSRFSuite) TestAuthenticityTokenIsValidInHeader() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	encCSRFToken, _ := mdwCSRFSecureCookie.Encode(s.config.HTTPCSRFCookieName, csrfToken)
	authenticityToken, _ := generateAuthenticityToken(csrfToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encCSRFToken})
	c.Request.Header.Set(s.config.HTTPCSRFRequestHeader, authenticityToken)
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *mdwCSRFSuite) TestAuthenticityTokenIsValidInPostFormValue() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	encCSRFToken, _ := mdwCSRFSecureCookie.Encode(s.config.HTTPCSRFCookieName, csrfToken)
	authenticityToken, _ := generateAuthenticityToken(csrfToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header:   map[string][]string{},
		Method:   "POST",
		PostForm: url.Values{},
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encCSRFToken})
	c.Request.PostForm.Add(mdwCSRFAuthenticityTemplateFieldName(c), authenticityToken)
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *mdwCSRFSuite) TestAuthenticityTokenIsValidInMultipartForm() {
	csrfToken, _ := generateRandomBytes(mdwCSRFTokenLength)
	encCSRFToken, _ := mdwCSRFSecureCookie.Encode(s.config.HTTPCSRFCookieName, csrfToken)
	authenticityToken, _ := generateAuthenticityToken(csrfToken)
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
		Method: "POST",
		MultipartForm: &multipart.Form{
			Value: map[string][]string{},
		},
	}
	c.Request.AddCookie(&http.Cookie{Name: s.config.HTTPCSRFCookieName, Value: encCSRFToken})
	c.Request.MultipartForm.Value[mdwCSRFAuthenticityTemplateFieldName(c)] = []string{authenticityToken}
	mdwCSRFHandler(c, s.config, s.logger)

	s.Equal(http.StatusOK, c.Writer.Status())
}

func (s *mdwCSRFSuite) TestCSRFSkipCheck() {
	c, _ := NewTestContext(s.recorder)
	_, exists := c.Get(mdwCSRFSkipCheckCtxKey.String())
	s.Equal(false, exists)

	c, _ = NewTestContext(s.recorder)
	CSRFSkipCheck()(c)
	_, exists = c.Get(mdwCSRFSkipCheckCtxKey.String())
	s.Equal(true, exists)
}

func (s *mdwCSRFSuite) TestCSRFTemplateField() {
	c, _ := NewTestContext(s.recorder)
	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "test")
	s.Equal(`<input type="hidden" name="authenticity_token" value="test">`, c.CSRFAuthenticityTemplateField())

	c, _ = NewTestContext(s.recorder)
	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "")
	s.Equal(`<input type="hidden" name="authenticity_token" value="">`, c.CSRFAuthenticityTemplateField())

	c, _ = NewTestContext(s.recorder)
	c.Set(mdwCSRFAuthenticityFieldNameCtxKey.String(), "my_authenticity_token")
	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "test")
	s.Equal(`<input type="hidden" name="my_authenticity_token" value="test">`, c.CSRFAuthenticityTemplateField())
}

func (s *mdwCSRFSuite) TestCSRFAuthenticityToken() {
	c, _ := NewTestContext(s.recorder)
	s.Equal("", c.CSRFAuthenticityToken())

	c, _ = NewTestContext(s.recorder)
	c.Set(mdwCSRFAuthenticityTokenCtxKey.String(), "test")
	s.Equal("test", c.CSRFAuthenticityToken())
}

func (s *mdwCSRFSuite) TestCSRFMiddleware() {
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	mdwCSRF(s.config, s.logger)(c)
	s.NotNil(mdwCSRFSecureCookie)
}

func TestMdwCSRFSuite(t *testing.T) {
	test.Run(t, new(mdwCSRFSuite))
}
