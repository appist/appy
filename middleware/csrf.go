package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

// CSRF token length in bytes.
const csrfTokenLength = 32

var (
	csrfSecureCookie    *securecookie.SecureCookie
	csrfCtxFieldNameKey = "appy.csrfFieldName"
	csrfCtxSkipCheckKey = "appy.csrfSkip"
	csrfCtxTokenKey     = "appy.csrfToken"
	csrfSafeMethods     = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	errCsrfNoReferer    = errors.New("the request referer is missing")
	errCsrfBadReferer   = errors.New("the request referer is invalid")
	errCsrfNoToken      = errors.New("the CSRF token is missing")
	errCsrfBadToken     = errors.New("the CSRF token is invalid")
	generateRandomBytes = func(n int) ([]byte, error) {
		b := make([]byte, n)
		_, err := rand.Read(b)
		if err != nil {
			return nil, err
		}

		return b, nil
	}
)

// CSRF is a middleware that provides Cross-Site Request Forgery protection.
func CSRF(config *support.ConfigT) gin.HandlerFunc {
	csrfSecureCookie = securecookie.New(config.HTTPCSRFSecret, nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(config.HTTPCSRFCookieMaxAge)

	return func(c *gin.Context) {
		csrfHandler(c, config)
	}
}

func csrfHandler(c *gin.Context, config *support.ConfigT) {
	if IsAPIOnly(c) == true {
		c.Set(csrfCtxSkipCheckKey, true)
	}

	skipCheck, exists := c.Get(csrfCtxSkipCheckKey)
	if exists == true && skipCheck == true {
		c.Next()
		return
	}

	realToken, err := getCSRFTokenFromCookie(c, config)
	if err != nil || len(realToken) != csrfTokenLength {
		realToken, err = generateRandomBytes(csrfTokenLength)
		if err != nil {
			support.Logger.Error(err)
			c.AbortWithError(http.StatusForbidden, err)
			return
		}

		err = saveCSRFTokenIntoCookie(realToken, c, config)
		if err != nil {
			support.Logger.Error(err)
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
	}

	c.Set(csrfCtxTokenKey, getCSRFMaskedToken(realToken))
	c.Set(csrfCtxFieldNameKey, strings.ToLower(config.HTTPCSRFFieldName))

	r := c.Request
	if !support.ArrayContains(csrfSafeMethods, r.Method) {
		// Enforce an origin check for HTTPS connections. As per the Django CSRF implementation (https://goo.gl/vKA7GE)
		// the Referer header is almost always present for same-domain HTTP requests.
		if r.TLS != nil {
			referer, err := url.Parse(r.Referer())
			if err != nil || referer.String() == "" {
				support.Logger.Error(errCsrfNoReferer)
				c.AbortWithError(http.StatusForbidden, errCsrfNoReferer)
				return
			}

			if !(referer.Scheme == "https" && referer.Host == r.Host) {
				support.Logger.Error(errCsrfBadReferer)
				c.AbortWithError(http.StatusForbidden, errCsrfBadReferer)
				return
			}
		}

		if realToken == nil {
			support.Logger.Error(errCsrfNoToken)
			c.AbortWithError(http.StatusForbidden, errCsrfNoToken)
			return
		}

		authenticityToken := getCSRFUnmaskedToken(getCSRFTokenFromRequest(c, config))
		if !compareTokens(authenticityToken, realToken) {
			support.Logger.Error(errCsrfBadToken)
			c.AbortWithError(http.StatusForbidden, errCsrfBadToken)
			return
		}
	}

	c.Writer.Header().Add("Vary", "Cookie")
	c.Next()
}

// CSRFSkipCheck skips the CSRF check for the request.
func CSRFSkipCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(csrfCtxSkipCheckKey, true)
		c.Next()
	}
}

// CSRFTemplateField is a template helper for html/template that provides an <input> field populated with a CSRF token.
func CSRFTemplateField(c *gin.Context) string {
	fieldName := csrfTemplateFieldName(c)

	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, CSRFToken(c))
}

// CSRFToken returns the CSRF token for the request.
func CSRFToken(c *gin.Context) string {
	val, exists := c.Get(csrfCtxTokenKey)
	if exists == true {
		if token, ok := val.(string); ok {
			return token
		}
	}

	return ""
}

func csrfTemplateFieldName(c *gin.Context) string {
	fieldName, exists := c.Get(csrfCtxFieldNameKey)

	if fieldName == "" || !exists {
		fieldName = "authenticity_token"
	}

	return strings.ToLower(fieldName.(string))
}

func compareTokens(a, b []byte) bool {
	// This is required as subtle.ConstantTimeCompare does not check for equal
	// lengths in Go versions prior to 1.3.
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(a, b) == 1
}

func getCSRFTokenFromCookie(c *gin.Context, config *support.ConfigT) ([]byte, error) {
	encodedToken, err := c.Cookie(config.HTTPCSRFCookieName)
	if err != nil {
		return nil, err
	}

	token := make([]byte, csrfTokenLength)
	err = csrfSecureCookie.Decode(config.HTTPCSRFCookieName, encodedToken, &token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getCSRFTokenFromRequest(c *gin.Context, config *support.ConfigT) []byte {
	r := c.Request
	fieldName := csrfTemplateFieldName(c)

	// 1. Check the HTTP header first.
	issued := r.Header.Get(http.CanonicalHeaderKey(config.HTTPCSRFRequestHeader))

	// 2. Fallback to the POST (form) value.
	if issued == "" {
		issued = r.PostFormValue(fieldName)
	}

	// 3. Finally, fallback to the multipart form (if set).
	if issued == "" && r.MultipartForm != nil {
		vals := r.MultipartForm.Value[fieldName]

		if len(vals) > 0 {
			issued = vals[0]
		}
	}

	// Decode the "issued" (pad + masked) token sent in the request. Return a nil byte slice on a decoding error.
	decoded, err := base64.StdEncoding.DecodeString(issued)
	if err != nil {
		return nil
	}

	return decoded
}

// getCSRFMaskedToken returns a unique-per-request token to mitigate the BREACH attack
// as per http://breachattack.com/#mitigations
//
// The token is generated by XOR'ing a one-time-pad and the base (session) CSRF
// token and returning them together as a 64-byte slice. This effectively
// randomises the token on a per-request basis without breaking multiple browser
// tabs/windows.
func getCSRFMaskedToken(realToken []byte) string {
	otp, err := generateRandomBytes(csrfTokenLength)
	if err != nil {
		return ""
	}

	// XOR the OTP with the real token to generate a masked token. Append the
	// OTP to the front of the masked token to allow unmasking in the subsequent
	// request.
	return base64.StdEncoding.EncodeToString(append(otp, xorToken(otp, realToken)...))
}

// getCSRFUnmaskedToken splits the issued token (one-time-pad + masked token) and returns the
// unmasked request token for comparison.
func getCSRFUnmaskedToken(issued []byte) []byte {
	// Issued tokens are always masked and combined with the pad.
	if len(issued) != csrfTokenLength*2 {
		return nil
	}

	// We now know the length of the byte slice.
	otp := issued[csrfTokenLength:]
	masked := issued[:csrfTokenLength]

	// Unmask the token by XOR'ing it against the OTP used to mask it.
	return xorToken(otp, masked)
}

// xorToken XORs tokens ([]byte) to provide unique-per-request CSRF tokens. It
// will return a masked token if the base token is XOR'ed with a one-time-pad.
// An unmasked token will be returned if a masked token is XOR'ed with the
// one-time-pad used to mask it.
func xorToken(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	res := make([]byte, n)

	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}

	return res
}

func saveCSRFTokenIntoCookie(token []byte, c *gin.Context, config *support.ConfigT) error {
	encoded, err := csrfSecureCookie.Encode(config.HTTPCSRFCookieName, token)
	if err != nil {
		return err
	}

	c.SetCookie(
		config.HTTPCSRFCookieName,
		encoded,
		config.HTTPCSRFCookieMaxAge,
		config.HTTPCSRFCookiePath,
		config.HTTPCSRFCookieDomain,
		config.HTTPCSRFCookieSecure,
		config.HTTPCSRFCookieHTTPOnly,
	)

	return nil
}
