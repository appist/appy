package http

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	appysupport "github.com/appist/appy/internal/support"
	"github.com/gorilla/securecookie"
)

// CSRF token length in bytes.
const csrfTokenLength = 32

var (
	csrfSecureCookie    *securecookie.SecureCookie
	csrfFieldNameCtxKey = ContextKey("csrfFieldName")
	csrfSkipCheckCtxKey = ContextKey("csrfSkipCheck")
	csrfTokenCtxKey     = ContextKey("csrfToken")
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
func CSRF(config *appysupport.Config, logger *appysupport.Logger) HandlerFunc {
	csrfSecureCookie = securecookie.New(config.HTTPCSRFSecret, nil)
	csrfSecureCookie.SetSerializer(securecookie.JSONEncoder{})
	csrfSecureCookie.MaxAge(config.HTTPCSRFCookieMaxAge)

	return func(c *Context) {
		csrfHandler(c, config, logger)
	}
}

func csrfHandler(ctx *Context, config *appysupport.Config, logger *appysupport.Logger) {
	if IsAPIOnly(ctx) == true {
		ctx.Set(csrfSkipCheckCtxKey.String(), true)
	}

	skipCheck, exists := ctx.Get(csrfSkipCheckCtxKey.String())
	if exists == true && skipCheck == true {
		ctx.Next()
		return
	}

	realToken, err := getCSRFTokenFromCookie(ctx, config)
	if err != nil || len(realToken) != csrfTokenLength {
		realToken, err = generateRandomBytes(csrfTokenLength)
		if err != nil {
			logger.Error(err)
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		err = saveCSRFTokenIntoCookie(realToken, ctx, config)
		if err != nil {
			logger.Error(err)
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}
	}

	authenticityToken := getCSRFMaskedToken(realToken)
	err = saveAuthenticityTokenIntoCookie(authenticityToken, ctx, config)
	if err != nil {
		logger.Error(err)
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	}

	ctx.Set(csrfTokenCtxKey.String(), authenticityToken)
	ctx.Set(csrfFieldNameCtxKey.String(), strings.ToLower(config.HTTPCSRFFieldName))

	r := ctx.Request
	if !appysupport.ArrayContains(csrfSafeMethods, r.Method) {
		// Enforce an origin check for HTTPS connections. As per the Django CSRF implementation (https://goo.gl/vKA7GE)
		// the Referer header is almost always present for same-domain HTTP requests.
		if r.TLS != nil {
			referer, err := url.Parse(r.Referer())
			if err != nil || referer.String() == "" {
				logger.Error(errCsrfNoReferer)
				ctx.AbortWithError(http.StatusForbidden, errCsrfNoReferer)
				return
			}

			if !(referer.Scheme == "https" && referer.Host == r.Host) {
				logger.Error(errCsrfBadReferer)
				ctx.AbortWithError(http.StatusForbidden, errCsrfBadReferer)
				return
			}
		}

		if realToken == nil {
			logger.Error(errCsrfNoToken)
			ctx.AbortWithError(http.StatusForbidden, errCsrfNoToken)
			return
		}

		authenticityToken := getCSRFUnmaskedToken(getCSRFTokenFromRequest(ctx, config))
		if !compareTokens(authenticityToken, realToken) {
			logger.Error(errCsrfBadToken)
			ctx.AbortWithError(http.StatusForbidden, errCsrfBadToken)
			return
		}
	}

	ctx.Writer.Header().Add("Vary", "Cookie")
	ctx.Next()
}

// CSRFSkipCheck skips the CSRF check for the request.
func CSRFSkipCheck() HandlerFunc {
	return func(ctx *Context) {
		ctx.Set(csrfSkipCheckCtxKey.String(), true)
		ctx.Next()
	}
}

// CSRFTemplateField is a template helper for html/template that provides an <input> field populated with a CSRF token.
func CSRFTemplateField(c *Context) string {
	fieldName := csrfTemplateFieldName(c)

	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, CSRFToken(c))
}

// CSRFToken returns the CSRF token for the request.
func CSRFToken(ctx *Context) string {
	val, exists := ctx.Get(csrfTokenCtxKey.String())
	if exists == true {
		if token, ok := val.(string); ok {
			return token
		}
	}

	return ""
}

func csrfTemplateFieldName(ctx *Context) string {
	fieldName, exists := ctx.Get(csrfFieldNameCtxKey.String())

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

func getCSRFTokenFromCookie(ctx *Context, config *appysupport.Config) ([]byte, error) {
	encodedToken, err := ctx.Cookie(config.HTTPCSRFCookieName)
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

func getCSRFTokenFromRequest(ctx *Context, config *appysupport.Config) []byte {
	r := ctx.Request
	fieldName := csrfTemplateFieldName(ctx)

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

func saveCSRFTokenIntoCookie(token []byte, ctx *Context, config *appysupport.Config) error {
	encoded, err := csrfSecureCookie.Encode(config.HTTPCSRFCookieName, token)
	if err != nil {
		return err
	}

	ctx.SetCookie(
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

func saveAuthenticityTokenIntoCookie(token string, ctx *Context, config *appysupport.Config) error {
	ctx.SetCookie(
		config.HTTPCSRFFieldName,
		token,
		config.HTTPCSRFCookieMaxAge,
		config.HTTPCSRFCookiePath,
		config.HTTPCSRFCookieDomain,
		config.HTTPCSRFCookieSecure,
		false,
	)

	return nil
}
