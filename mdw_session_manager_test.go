package appy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ginsessions "github.com/gin-contrib/sessions"
)

type SessionManagerSuite struct {
	TestSuite
	asset    *Asset
	config   *Config
	logger   *Logger
	recorder *httptest.ResponseRecorder
}

func testSessionOps(s *SessionManagerSuite, session Sessioner) {
	session.AddFlash("i am a bot", "message")
	s.Nil(session.Save())

	flashes := session.Flashes("message")
	s.Equal(1, len(flashes))
	s.Equal("i am a bot", flashes[0])

	expected := map[interface{}]interface{}{
		"baby": "i am a baby",
		"bot":  "i am a bot",
		"cat":  "i am a cat",
	}

	for key, val := range expected {
		session.Set(key, val)
	}

	s.Nil(session.Save())
	s.Equal(expected, session.Values())

	session.Delete("baby")
	s.Nil(session.Get("baby"))

	session.Clear()
	s.Nil(session.Get("cat"))
}

func (s *SessionManagerSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_REDIS_ADDR", "0.0.0.0:16379")

	s.logger = NewLogger()
	s.asset = NewAsset(http.Dir("testdata"), nil, "")
	s.config = NewConfig(s.asset, s.logger, &Support{})
	s.recorder = httptest.NewRecorder()
}

func (s *SessionManagerSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *SessionManagerSuite) TestSessionUnknownStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}

	s.config.HTTPSessionProvider = "unknown"
	s.Panics(func() { SessionManager(s.config)(c) })
}

func (s *SessionManagerSuite) TestSessionCookieStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "cookie"
	SessionManager(s.config)(c)

	session := c.Session()
	session.Options(ginsessions.Options{
		MaxAge: 368400,
	})
	sessionCookie, _ := c.Cookie(s.config.HTTPSessionName)
	s.NotNil(sessionCookie)
	testSessionOps(s, session)
}

func (s *SessionManagerSuite) TestSessionRedisStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	SessionManager(s.config)(c)

	session := c.Session()
	testSessionOps(s, session)
}

func (s *SessionManagerSuite) TestSessionRedisStoreWrongAddr() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAddr = "localhost:1234"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func (s *SessionManagerSuite) TestSessionRedisStoreWrongAuth() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAuth = "authme"
	s.Panics(func() { SessionManager(s.config)(c) })
}

func (s *SessionManagerSuite) TestSessionRedisStoreInvalidDb() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisDb = "-1"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func (s *SessionManagerSuite) TestNonExistentSession() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.Nil(c.Session())
}

func (s *SessionManagerSuite) TestCustomSessionKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	SessionManager(s.config)(c)

	session := c.Session()
	session.SetKeyPrefix("mysession:")
	testSessionOps(s, session)

	s.Contains(session.Key(), "mysession:")
}

func TestSessionManagerSuite(t *testing.T) {
	RunTestSuite(t, new(SessionManagerSuite))
}
