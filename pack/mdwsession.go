package pack

import (
	"fmt"
	"net/http"

	"github.com/appist/appy/pack/internal/sessionstore"
	"github.com/appist/appy/support"
	ginsessions "github.com/gin-contrib/sessions"
	"github.com/go-redis/redis/v7"
	gorcontext "github.com/gorilla/context"
	gorsessions "github.com/gorilla/sessions"
)

var (
	mdwSessionCtxKey = ContextKey("sessionManager")
)

// SessionOptions defines the session cookie's configuration.
type SessionOptions = ginsessions.Options

// Sessioner stores the values and optional configuration for a session.
type Sessioner interface {
	// AddFlash adds a flash message to the session.
	// A single variadic argument is accepted, and it is optional: it defines the flash key.
	// If not defined "_flash" is used by default.
	AddFlash(value interface{}, vars ...string)

	// Clear deletes all values in the session.
	Clear()

	// Delete removes the session value associated to the given key.
	Delete(key interface{})

	// Flashes returns a slice of flash messages from the session.
	// A single variadic argument is accepted, and it is optional: it defines the flash key.
	// If not defined "_flash" is used by default.
	Flashes(vars ...string) []interface{}

	// Get returns the session value associated to the given key.
	Get(key interface{}) interface{}

	// Key returns the session key.
	Key() string

	// Options sets the cookie configuration for a session.
	Options(SessionOptions)

	// Set sets the session value associated to the given key.
	Set(key interface{}, val interface{})

	// Save saves all sessions used during the current request.
	Save() error

	// KeyPrefix returns the key prefix for the session, not available for CookieStore.
	KeyPrefix() string

	// SetKeyPrefix sets the key prefix for the session, not available for CookieStore.
	SetKeyPrefix(p string)

	// Values returns all values in the session.
	Values() map[interface{}]interface{}
}

func mdwSession(config *support.Config) HandlerFunc {
	return func(c *Context) {
		sessionStore, err := newSessionStore(config)
		if err != nil {
			panic(err)
		}

		s := &Session{config.HTTPSessionCookieName, c.Request, sessionStore, nil, false, c.Writer}
		c.Set(mdwSessionCtxKey.String(), s)
		defer gorcontext.Clear(c.Request)
		c.Next()
	}
}

// SessionStore is an interface for custom session stores.
type SessionStore interface {
	gorsessions.Store

	// Options sets the cookie configuration for a session.
	Options(SessionOptions)

	// KeyPrefix returns the prefix for the store key, not available for CookieStore.
	KeyPrefix() string

	// SetKeyPrefix sets the prefix for the store key, not available for CookieStore.
	SetKeyPrefix(p string)
}

func newSessionStore(config *support.Config) (SessionStore, error) {
	var (
		sessionStore SessionStore
		err          error
	)

	switch provider := config.HTTPSessionProvider; provider {
	case "cookie":
		sessionStore = sessionstore.NewCookieStore(config.HTTPSessionSecrets...)
	case "redis":
		sessionStore, err = sessionstore.NewRedisStore(&redis.Options{
			Addr:               config.HTTPSessionRedisAddr,
			Password:           config.HTTPSessionRedisPassword,
			DB:                 config.HTTPSessionRedisDB,
			IdleCheckFrequency: config.HTTPSessionRedisIdleCheckFrequency,
			IdleTimeout:        config.HTTPSessionRedisIdleTimeout,
			MaxConnAge:         config.HTTPSessionRedisMaxConnAge,
			MinIdleConns:       config.HTTPSessionRedisMinIdleConns,
			PoolSize:           config.HTTPSessionRedisPoolSize,
			PoolTimeout:        config.HTTPSessionRedisPoolTimeout,
		}, config.HTTPSessionSecrets...)
	default:
		err = fmt.Errorf("session provider '%s' is not supported", provider)
	}

	if sessionStore != nil {
		sessionStore.Options(SessionOptions{
			Domain:   config.HTTPSessionCookieDomain,
			HttpOnly: config.HTTPSessionCookieHTTPOnly,
			MaxAge:   config.HTTPSessionExpiration,
			Path:     config.HTTPSessionCookiePath,
			Secure:   config.HTTPSessionCookieSecure,
			SameSite: config.HTTPSessionCookieSameSite,
		})
	}

	return sessionStore, err
}

// Session provides the session functionality for a single HTTP request.
type Session struct {
	name    string
	request *http.Request
	store   SessionStore
	session *gorsessions.Session
	written bool
	writer  http.ResponseWriter
}

// AddFlash adds a flash message to the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *Session) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

// Get returns the session value associated to the given key.
func (s *Session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

// Delete removes the session value associated to the given key.
func (s *Session) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

// Clear deletes all values in the session.
func (s *Session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

// Flashes returns a slice of flash messages from the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *Session) Flashes(vars ...string) []interface{} {
	s.written = true
	return s.Session().Flashes(vars...)
}

// Options sets configuration for a session.
func (s *Session) Options(options SessionOptions) {
	s.Session().Options = &gorsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		SameSite: options.SameSite,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}

// Save saves all sessions used during the current request.
func (s *Session) Save() error {
	if !s.Written() {
		return nil
	}

	err := s.Session().Save(s.request, s.writer)
	if err == nil {
		s.written = false
	}

	return err
}

// Session retrieves the data for the specific HTTP request.
func (s *Session) Session() *gorsessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.name)

		if err != nil {
			return nil
		}
	}

	return s.session
}

// Key returns the session key.
func (s *Session) Key() string {
	return s.KeyPrefix() + s.session.ID
}

// KeyPrefix returns the key prefix for the session, not available for CookieStore.
func (s *Session) KeyPrefix() string {
	return s.store.KeyPrefix()
}

// SetKeyPrefix sets the key prefix for the session, not available for CookieStore.
func (s *Session) SetKeyPrefix(p string) {
	s.store.SetKeyPrefix(p)
}

// Set sets the session value associated to the given key.
func (s *Session) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

// Values returns all values in the session.
func (s *Session) Values() map[interface{}]interface{} {
	if s.Session() == nil {
		return nil
	}

	return s.Session().Values
}

// Written indicates if the session's key/value map had already been stored into the data store.
func (s *Session) Written() bool {
	return s.written
}
