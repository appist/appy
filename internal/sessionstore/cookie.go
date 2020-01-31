package sessionstore

import (
	ginsessions "github.com/gin-contrib/sessions"
	gorsessions "github.com/gorilla/sessions"
)

// CookieStore stores sessions in the browser cookie.
type CookieStore struct {
	*gorsessions.CookieStore
}

// NewCookieStore initializes a CookieStore.
func NewCookieStore(keyPairs ...[]byte) Store {
	return &CookieStore{gorsessions.NewCookieStore(keyPairs...)}
}

// Options defines how the session cookie should be configured.
func (s *CookieStore) Options(options ginsessions.Options) {
	s.CookieStore.Options = &gorsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}

// KeyPrefix doesn't do anything for cookie store.
func (s *CookieStore) KeyPrefix() string {
	return ""
}

// SetKeyPrefix doesn't do anything for cookie store.
func (s *CookieStore) SetKeyPrefix(p string) {
}
