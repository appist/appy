package sessionstore

import (
	ginsessions "github.com/gin-contrib/sessions"
)

type (
	// Store provides an interface to implement various stores.
	Store interface {
		ginsessions.Store

		// SetKeyPrefix returns the prefix for the store key, not available for CookieStore.
		KeyPrefix() string

		// SetKeyPrefix sets the prefix for the store key, not available for CookieStore.
		SetKeyPrefix(p string)
	}
)
