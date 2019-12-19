package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	ginsessions "github.com/gin-contrib/sessions"
)

type SessionMngrSuite struct {
	test.Suite
	assets   *support.Assets
	config   *support.Config
	logger   *support.Logger
	recorder *httptest.ResponseRecorder
}

func testSessionOps(s *SessionMngrSuite, session Sessioner) {
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

func (s *SessionMngrSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger = support.NewLogger()
	s.assets = support.NewAssets(nil, "", http.Dir("../support/testdata"))
	s.config = support.NewConfig(s.assets, s.logger)
	s.recorder = httptest.NewRecorder()
}

func (s *SessionMngrSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *SessionMngrSuite) TestSessionUnknownStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}

	s.config.HTTPSessionProvider = "unknown"
	s.Panics(func() { SessionMngr(s.config)(c) })
}

func (s *SessionMngrSuite) TestSessionCookieStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "cookie"
	SessionMngr(s.config)(c)

	session := DefaultSession(c)
	session.Options(ginsessions.Options{
		MaxAge: 368400,
	})
	sessionCookie, _ := c.Cookie(s.config.HTTPSessionName)
	s.NotNil(sessionCookie)
	testSessionOps(s, session)
}

func (s *SessionMngrSuite) TestSessionRedisStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	SessionMngr(s.config)(c)

	session := DefaultSession(c)
	testSessionOps(s, session)
}

func (s *SessionMngrSuite) TestSessionRedisStoreWrongAddr() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAddr = "localhost:1234"
	s.Panics(func() { SessionMngr(s.config)(ctx) })
}

func (s *SessionMngrSuite) TestSessionRedisStoreWrongAuth() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAuth = "authme"
	s.Panics(func() { SessionMngr(s.config)(c) })
}

func (s *SessionMngrSuite) TestSessionRedisStoreInvalidDb() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisDb = "-1"
	s.Panics(func() { SessionMngr(s.config)(ctx) })
}

func (s *SessionMngrSuite) TestNonExistentDefaultSession() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.Nil(DefaultSession(c))
}

func (s *SessionMngrSuite) TestCustomSessionKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	SessionMngr(s.config)(c)

	session := DefaultSession(c)
	session.SetKeyPrefix("mysession:")
	testSessionOps(s, session)

	s.Contains(session.Key(), "mysession:")
}

func TestSessionManager(t *testing.T) {
	test.RunSuite(t, new(SessionMngrSuite))
}
