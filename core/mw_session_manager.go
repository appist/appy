package core

import (
	"fmt"
	"net/http"

	ginsessions "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	rr "github.com/gomodule/redigo/redis"
	"github.com/gorilla/context"
	gorsessions "github.com/gorilla/sessions"
)

var (
	sessionManagerCtxKey = "appy.session"
)

// SessionStore is an interface for custom session stores.
type SessionStore interface {
	gorsessions.Store
	Options(ginsessions.Options)
}

// Session stores the values and optional configuration for a session.
type Session interface {
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
	// Options sets configuration for a session.
	Options(ginsessions.Options)
	// Set sets the session value associated to the given key.
	Set(key interface{}, val interface{})
	// Save saves all sessions used during the current request.
	Save() error
	// Values returns all values in the session.
	Values() map[interface{}]interface{}
}

// AppSession provides the session functionality for a single HTTP request.
type AppSession struct {
	name    string
	request *http.Request
	store   SessionStore
	session *gorsessions.Session
	written bool
	writer  http.ResponseWriter
}

// SessionManager is a middleware that provides the session management functionality.
func SessionManager(config AppConfig) HandlerFunc {
	return func(ctx *Context) {
		sessionStore, err := newSessionStore(config)
		if err != nil {
			panic(err)
		}

		s := &AppSession{config.HTTPSessionName, ctx.Request, sessionStore, nil, false, ctx.Writer}
		ctx.Set(sessionManagerCtxKey, s)
		defer context.Clear(ctx.Request)
		ctx.Next()
	}
}

func newSessionStore(config AppConfig) (SessionStore, error) {
	var (
		sessionStore SessionStore
		err          error
	)

	switch provider := config.HTTPSessionProvider; provider {
	case "cookie":
		sessionStore = cookie.NewStore(config.HTTPSessionSecrets...)
		sessionStore.Options(ginsessions.Options{
			Domain:   config.HTTPSessionCookieDomain,
			HttpOnly: config.HTTPSessionCookieHTTPOnly,
			MaxAge:   config.HTTPSessionCookieMaxAge,
			Path:     config.HTTPSessionCookiePath,
			Secure:   config.HTTPSessionCookieSecure,
		})
	case "redis":
		sessionStore, err = redis.NewStoreWithPool(
			newSessionRedisPool(config),
			config.HTTPSessionSecrets...,
		)
	default:
		err = fmt.Errorf("session provider '%s' is not supported", provider)
	}

	return sessionStore, err
}

func newSessionRedisPool(config AppConfig) *rr.Pool {
	return &rr.Pool{
		Dial: func() (rr.Conn, error) {
			conn, err := rr.Dial("tcp", config.HTTPSessionRedisAddr)
			if err != nil {
				return nil, err
			}

			if config.HTTPSessionRedisAuth != "" {
				if _, err := conn.Do("AUTH", config.HTTPSessionRedisAuth); err != nil {
					conn.Close()
					return nil, err
				}
			}

			if _, err := conn.Do("SELECT", config.HTTPSessionRedisDb); err != nil {
				conn.Close()
				return nil, err
			}

			return conn, nil
		},
		IdleTimeout:     config.HTTPSessionRedisIdleTimeout,
		MaxConnLifetime: config.HTTPSessionRedisMaxConnLifetime,
		MaxActive:       config.HTTPSessionRedisMaxActive,
		MaxIdle:         config.HTTPSessionRedisMaxIdle,
		Wait:            config.HTTPSessionRedisWait,
	}
}

// AddFlash adds a flash message to the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *AppSession) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

// Get returns the session value associated to the given key.
func (s *AppSession) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

// Delete removes the session value associated to the given key.
func (s *AppSession) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

// Clear deletes all values in the session.
func (s *AppSession) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

// Flashes returns a slice of flash messages from the session.
// A single variadic argument is accepted, and it is optional: it defines the flash key.
// If not defined "_flash" is used by default.
func (s *AppSession) Flashes(vars ...string) []interface{} {
	s.written = true
	return s.Session().Flashes(vars...)
}

// Options sets configuration for a session.
func (s *AppSession) Options(options ginsessions.Options) {
	s.Session().Options = &gorsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}

// Save saves all sessions used during the current request.
func (s *AppSession) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.writer)
		if e == nil {
			s.written = false
		}

		return e
	}

	return nil
}

// Session retrieves the data for the specific HTTP request.
func (s *AppSession) Session() *gorsessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.name)
		if err != nil {
			return nil
		}
	}

	return s.session
}

// Set sets the session value associated to the given key.
func (s *AppSession) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

// Values returns all values in the session.
func (s *AppSession) Values() map[interface{}]interface{} {
	return s.Session().Values
}

// Written indicates if the session's key/value map had already been stored into the data store.
func (s *AppSession) Written() bool {
	return s.written
}

// DefaultSession returns the session in the request context.
func DefaultSession(c *Context) Session {
	s, exists := c.Get(sessionManagerCtxKey)
	if !exists {
		return nil
	}

	return s.(Session)
}
