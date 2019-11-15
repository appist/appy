package appy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/appist/appy/sessionstore"
	ginsessions "github.com/gin-contrib/sessions"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/context"
	gorsessions "github.com/gorilla/sessions"
)

var (
	sessionManagerCtxKey = ContextKey("sessionManager")
)

// SessionStore is an interface for custom session stores.
type SessionStore interface {
	gorsessions.Store

	// Options sets the cookie configuration for a session.
	Options(ginsessions.Options)

	// KeyPrefix returns the prefix for the store key, not available for CookieStore.
	KeyPrefix() string

	// SetKeyPrefix sets the prefix for the store key, not available for CookieStore.
	SetKeyPrefix(p string)
}

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
	Options(ginsessions.Options)

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

// Session provides the session functionality for a single HTTP request.
type Session struct {
	name    string
	request *http.Request
	store   SessionStore
	session *gorsessions.Session
	written bool
	writer  http.ResponseWriter
}

// RedisPoolConfig keeps the redis connection pool config.
type RedisPoolConfig struct {
	Addr, Auth, Db               string
	MaxActive, MaxIdle           int
	Wait                         bool
	IdleTimeout, MaxConnLifetime time.Duration
}

// SessionManager is a middleware that provides the session management functionality.
func SessionManager(config *Config) HandlerFunc {
	return func(ctx *Context) {
		sessionStore, err := newSessionStore(config)
		if err != nil {
			panic(err)
		}

		s := &Session{config.HTTPSessionName, ctx.Request, sessionStore, nil, false, ctx.Writer}
		ctx.Set(sessionManagerCtxKey.String(), s)
		defer context.Clear(ctx.Request)
		ctx.Next()
	}
}

func newSessionStore(config *Config) (SessionStore, error) {
	var (
		sessionStore SessionStore
		err          error
	)

	switch provider := config.HTTPSessionProvider; provider {
	case "cookie":
		sessionStore = sessionstore.NewCookieStore(config.HTTPSessionSecrets...)
	case "redis":
		redisPoolConfig := RedisPoolConfig{
			Addr:            config.HTTPSessionRedisAddr,
			Auth:            config.HTTPSessionRedisAuth,
			Db:              config.HTTPSessionRedisDb,
			IdleTimeout:     config.HTTPSessionRedisIdleTimeout,
			MaxConnLifetime: config.HTTPSessionRedisMaxConnLifetime,
			MaxActive:       config.HTTPSessionRedisMaxActive,
			MaxIdle:         config.HTTPSessionRedisMaxIdle,
			Wait:            config.HTTPSessionRedisWait,
		}

		sessionStore, err = sessionstore.NewRedisStoreWithPool(
			NewRedisPool(redisPoolConfig),
			config.HTTPSessionSecrets...,
		)
	default:
		err = fmt.Errorf("session provider '%s' is not supported", provider)
	}

	if sessionStore != nil {
		sessionStore.Options(ginsessions.Options{
			Domain:   config.HTTPSessionDomain,
			HttpOnly: config.HTTPSessionHTTPOnly,
			MaxAge:   config.HTTPSessionExpiration,
			Path:     config.HTTPSessionPath,
			Secure:   config.HTTPSessionSecure,
		})
	}

	return sessionStore, err
}

// NewRedisPool initializes the redis connection pool.
func NewRedisPool(config RedisPoolConfig) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", config.Addr)
			if err != nil {
				return nil, err
			}

			if config.Auth != "" {
				if _, err := conn.Do("AUTH", config.Auth); err != nil {
					conn.Close()
					return nil, err
				}
			}

			if _, err := conn.Do("SELECT", config.Db); err != nil {
				conn.Close()
				return nil, err
			}

			return conn, nil
		},
		IdleTimeout:     config.IdleTimeout,
		MaxConnLifetime: config.MaxConnLifetime,
		MaxActive:       config.MaxActive,
		MaxIdle:         config.MaxIdle,
		Wait:            config.Wait,
	}
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
func (s *Session) Options(options ginsessions.Options) {
	s.Session().Options = &gorsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}

// Save saves all sessions used during the current request.
func (s *Session) Save() error {
	if s.Written() {
		err := s.Session().Save(s.request, s.writer)
		if err == nil {
			s.written = false
		}

		return err
	}

	return nil
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
	return s.store.KeyPrefix() + s.session.ID
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
	return s.Session().Values
}

// Written indicates if the session's key/value map had already been stored into the data store.
func (s *Session) Written() bool {
	return s.written
}

// DefaultSession returns the session in the request context.
func DefaultSession(ctx *Context) Sessioner {
	s, exists := ctx.Get(sessionManagerCtxKey.String())
	if !exists {
		return nil
	}

	return s.(Sessioner)
}
