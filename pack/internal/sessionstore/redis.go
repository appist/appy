package sessionstore

import (
	"bytes"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"net/http"
	"strings"

	ginsessions "github.com/gin-contrib/sessions"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/securecookie"
	gorsessions "github.com/gorilla/sessions"
)

type (
	// SessionSerializer provides an interface hook for alternative serializers.
	SessionSerializer interface {
		Deserialize(d []byte, ss *gorsessions.Session) error
		Serialize(ss *gorsessions.Session) ([]byte, error)
	}

	// RedisStore stores sessions in the redis backend.
	RedisStore struct {
		redisClient   *redis.Client
		Codecs        []securecookie.Codec
		CookieOptions *gorsessions.Options // default configuration
		DefaultMaxAge int                  // default Redis TTL for a MaxAge == 0 session
		maxLength     int
		keyPrefix     string
		serializer    SessionSerializer
	}

	// GobSerializer uses gob package to encode the session map.
	GobSerializer struct{}
)

var (
	defaultCookieMaxAge = 86400 * 14
)

// Serialize using gob
func (s GobSerializer) Serialize(ss *gorsessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(ss.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

// Deserialize uses gob package to decode the session map.
func (s GobSerializer) Deserialize(d []byte, ss *gorsessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(&ss.Values)
}

// NewRedisStore initializes a RedisStore instance with connections pool.
func NewRedisStore(opts *redis.Options, keyPairs ...[]byte) (Store, error) {
	rs := &RedisStore{
		redisClient: redis.NewClient(opts),
		Codecs:      securecookie.CodecsFromPairs(keyPairs...),
		CookieOptions: &gorsessions.Options{
			Path:   "/",
			MaxAge: defaultCookieMaxAge,
		},
		DefaultMaxAge: 60 * 20, // 20 minutes seems like a reasonable default
		maxLength:     4096,
		keyPrefix:     "session:",
		serializer:    GobSerializer{},
	}

	_, err := rs.ping()
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// Close closes the underlying *redis.Pool.
func (s *RedisStore) Close() error {
	return s.redisClient.Close()
}

// Get returns a session for the given name after adding it to the registry.
func (s *RedisStore) Get(r *http.Request, name string) (*gorsessions.Session, error) {
	return gorsessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
func (s *RedisStore) New(r *http.Request, name string) (*gorsessions.Session, error) {
	var (
		err error
		ok  bool
	)
	session := gorsessions.NewSession(s, name)
	// make a copy
	options := *s.CookieOptions
	session.Options = &options
	session.IsNew = true
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			ok, err = s.load(session)
			session.IsNew = !(err == nil && ok) // not new if no error and data available
		}
	}

	return session, err
}

// Options defines how the session cookie should be configured.
func (s *RedisStore) Options(options ginsessions.Options) {
	s.CookieOptions = &gorsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		SameSite: options.SameSite,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
}

// Save adds a single session to the response.
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *gorsessions.Session) error {
	// Marked for deletion.
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}

		http.SetCookie(w, gorsessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	// Build an alphanumeric key for the redis store.
	if session.ID == "" {
		session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	}

	if err := s.save(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, gorsessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// KeyPrefix returns the prefix for the redis key.
func (s *RedisStore) KeyPrefix() string {
	return s.keyPrefix
}

// SetKeyPrefix sets the prefix for the redis key.
func (s *RedisStore) SetKeyPrefix(p string) {
	s.keyPrefix = p
}

// ping does an internal ping against a server to check if it is alive.
func (s *RedisStore) ping() (bool, error) {
	data, err := s.redisClient.Ping().Result()
	if err != nil || data == "" {
		return false, err
	}

	return (data == "PONG"), nil
}

// save stores the session in redis.
func (s *RedisStore) save(session *gorsessions.Session) error {
	b, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}

	if s.maxLength != 0 && len(b) > s.maxLength {
		return errors.New("the value to store into session is too big")
	}

	age := session.Options.MaxAge
	if age == 0 {
		age = s.DefaultMaxAge
	}

	_, err = s.redisClient.Do("SETEX", s.keyPrefix+session.ID, age, b).Result()
	if err != nil {
		return err
	}

	return nil
}

// load reads the session from redis and returns true if there is a sessoin data in DB.
func (s *RedisStore) load(session *gorsessions.Session) (bool, error) {
	data, err := s.redisClient.Do("GET", s.keyPrefix+session.ID).Result()
	if err != nil {
		return false, err
	}

	if data == nil {
		return false, nil // no data was associated with this key
	}

	b, ok := data.([]byte)
	if !ok {
		return false, nil
	}

	return true, s.serializer.Deserialize(b, session)
}

// delete removes keys from redis if MaxAge<0
func (s *RedisStore) delete(session *gorsessions.Session) error {
	if _, err := s.redisClient.Do("DEL", s.keyPrefix+session.ID).Result(); err != nil {
		return err
	}

	return nil
}
