package appy

import (
	"fmt"
	"net/http"
	"strings"
)

type (
	securePolicy struct {
		config       secureConfig
		fixedHeaders []secureHeader
	}

	secureHeader struct {
		key   string
		value string
	}

	secureConfig struct {
		// AllowedHosts is a list of fully qualified domain names that are allowed.
		// Default is empty list, which allows any and all host names.
		AllowedHosts []string

		// If SSLRedirect is set to true, then only allow https requests.
		// Default is false.
		SSLRedirect bool

		// If SSLTemporaryRedirect is true, the a 302 will be used while redirecting.
		// Default is false (301).
		SSLTemporaryRedirect bool

		// SSLHost is the host name that is used to redirect http requests to https.
		// Default is "", which indicates to use the same host.
		SSLHost string

		// STSSeconds is the max-age of the Strict-Transport-Security header.
		// Default is 0, which would NOT include the header.
		STSSeconds int64

		// If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the 
		// Strict-Transport-Security header. Default is false.
		STSIncludeSubdomains bool

		// If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
		FrameDeny bool

		// CustomFrameOptionsValue allows the X-Frame-Options header value to be set with a custom value. This 
		// overrides the FrameDeny option.
		CustomFrameOptionsValue string

		// If ContentTypeNosniff is true, adds the X-Content-Type-Options header
		// with the value `nosniff`. Default is false.
		ContentTypeNosniff bool

		// If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
		BrowserXSSFilter bool

		// ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. 
		// Default is "".
		ContentSecurityPolicy string

		// HTTP header "Referrer-Policy" governs which referrer information, sent in the Referrer header, should be 
		// included with requests made.
		ReferrerPolicy string

		// When true, the whole secury policy applied by the middleware is disable completely.
		IsDevelopment bool

		// Prevent Internet Explorer from executing downloads in your site’s context.
		IENoOpen bool

		// If the request is insecure, treat it as secure if any of the headers in this dict are set to their corresponding
		// value. This is useful when your app is running behind a secure proxy that forwards requests to your app over 
		// http (such as on Heroku).
		SSLProxyHeaders map[string]string
	}
)

// Secure is a middleware that provides security settings for the server.
func Secure(config *Config) HandlerFunc {
	policy := newSecurePolicy(newSecureConfig(config))

	return func(c *Context) {
		if !policy.applyToContext(c) {
			return
		}
	}
}

func newSecureConfig(config *Config) secureConfig {
	return secureConfig{
		IsDevelopment:           false,
		AllowedHosts:            config.HTTPAllowedHosts,
		SSLRedirect:             config.HTTPSSLRedirect,
		SSLTemporaryRedirect:    config.HTTPSSLTemporaryRedirect,
		SSLHost:                 config.HTTPSSLHost,
		STSSeconds:              config.HTTPSTSSeconds,
		STSIncludeSubdomains:    config.HTTPSTSIncludeSubdomains,
		FrameDeny:               config.HTTPFrameDeny,
		CustomFrameOptionsValue: config.HTTPCustomFrameOptionsValue,
		ContentTypeNosniff:      config.HTTPContentTypeNosniff,
		BrowserXSSFilter:        config.HTTPBrowserXSSFilter,
		ContentSecurityPolicy:   config.HTTPContentSecurityPolicy,
		ReferrerPolicy:          config.HTTPReferrerPolicy,
		IENoOpen:                config.HTTPIENoOpen,
		SSLProxyHeaders:         config.HTTPSSLProxyHeaders,
	}
}

func newSecurePolicy(config secureConfig) *securePolicy {
	policy := &securePolicy{}
	policy.loadConfig(config)

	return policy
}

func (p *securePolicy) loadConfig(config secureConfig) {
	p.config = config
	p.fixedHeaders = make([]secureHeader, 0, 5)

	// Frame Options header.
	if len(config.CustomFrameOptionsValue) > 0 {
		p.addHeader("X-Frame-Options", config.CustomFrameOptionsValue)
	} else if config.FrameDeny {
		p.addHeader("X-Frame-Options", "DENY")
	}

	// Content Type Options header.
	if config.ContentTypeNosniff {
		p.addHeader("X-Content-Type-Options", "nosniff")
	}

	// XSS Protection header.
	if config.BrowserXSSFilter {
		p.addHeader("X-XSS-Protection", "1; mode=block")
	}

	// Content Security Policy header.
	if len(config.ContentSecurityPolicy) > 0 {
		p.addHeader("Content-Security-Policy", config.ContentSecurityPolicy)
	}

	if len(config.ReferrerPolicy) > 0 {
		p.addHeader("Referrer-Policy", config.ReferrerPolicy)
	}

	// Strict Transport Security header.
	if config.STSSeconds != 0 {
		stsSub := ""
		if config.STSIncludeSubdomains {
			stsSub = "; includeSubdomains"
		}

		// TODO
		// "max-age=%d%s" refactor
		p.addHeader(
			"Strict-Transport-Security",
			fmt.Sprintf("max-age=%d%s", config.STSSeconds, stsSub))
	}

	// X-Download-Options header.
	if config.IENoOpen {
		p.addHeader("X-Download-Options", "noopen")
	}
}

func (p *securePolicy) addHeader(key string, value string) {
	p.fixedHeaders = append(p.fixedHeaders, secureHeader{
		key:   key,
		value: value,
	})
}

func (p *securePolicy) applyToContext(c *Context) bool {
	if !p.config.IsDevelopment {
		p.writeSecureHeaders(c)

		if !p.checkAllowHosts(c) {
			return false
		}

		if !p.checkSSL(c) {
			return false
		}
	}

	return true
}

func (p *securePolicy) writeSecureHeaders(c *Context) {
	header := c.Writer.Header()
	for _, pair := range p.fixedHeaders {
		header.Set(pair.key, pair.value)
	}
}

func (p *securePolicy) checkAllowHosts(c *Context) bool {
	if len(p.config.AllowedHosts) == 0 {
		return true
	}

	host := c.Request.Host
	if len(host) == 0 {
		host = c.Request.URL.Host
	}

	for _, allowedHost := range p.config.AllowedHosts {
		if strings.EqualFold(allowedHost, host) {
			return true
		}
	}

	c.AbortWithStatus(http.StatusForbidden)
	return false
}

func (p *securePolicy) isSSLRequest(req *http.Request) bool {
	if strings.EqualFold(req.URL.Scheme, "https") || req.TLS != nil {
		return true
	}

	for h, v := range p.config.SSLProxyHeaders {
		hv, ok := req.Header[h]

		if !ok {
			continue
		}

		if strings.EqualFold(hv[0], v) {
			return true
		}
	}

	return false
}

func (p *securePolicy) checkSSL(c *Context) bool {
	if !p.config.SSLRedirect {
		return true
	}

	req := c.Request
	isSSLRequest := p.isSSLRequest(req)
	if isSSLRequest {
		return true
	}

	// TODO
	// req.Host vs req.URL.Host
	url := req.URL
	url.Scheme = "https"
	url.Host = req.Host

	if len(p.config.SSLHost) > 0 {
		url.Host = p.config.SSLHost
	}

	status := http.StatusMovedPermanently
	if p.config.SSLTemporaryRedirect {
		status = http.StatusFound
	}

	c.Redirect(status, url.String())
	c.Abort()
	return false
}
