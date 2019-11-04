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
	s.logger = NewLogger(DebugBuild)
	s.config = NewConfig(DebugBuild, s.logger, nil)
	s.recorder = httptest.NewRecorder()
}

func (s *SessionManagerSuite) TearDownTest() {
	os.Clearenv()
}

func (s *SessionManagerSuite) TestSessionUnknownStore() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}

	s.config.HTTPSessionProvider = "unknown"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func (s *SessionManagerSuite) TestSessionCookieStore() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "cookie"
	SessionManager(s.config)(ctx)

	session := DefaultSession(ctx)
	session.Options(ginsessions.Options{
		MaxAge: 368400,
	})
	sessionCookie, _ := ctx.Cookie(s.config.HTTPSessionName)
	s.NotNil(sessionCookie)
	testSessionOps(s, session)
}

func (s *SessionManagerSuite) TestSessionRedisStore() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	SessionManager(s.config)(ctx)

	session := DefaultSession(ctx)
	testSessionOps(s, session)
}

func (s *SessionManagerSuite) TestSessionRedisStoreWrongAddr() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAddr = "localhost:1234"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func (s *SessionManagerSuite) TestSessionRedisStoreWrongAuth() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAuth = "authme"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func (s *SessionManagerSuite) TestSessionRedisStoreInvalidDb() {
	ctx, _ := CreateTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisDb = "-1"
	s.Panics(func() { SessionManager(s.config)(ctx) })
}

func TestSessionManager(t *testing.T) {
	RunTestSuite(t, new(SessionManagerSuite))
}
