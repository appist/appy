package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	ginsessions "github.com/gin-contrib/sessions"
)

type SessionManagerSuiteT struct {
	test.SuiteT
	Config   *support.ConfigT
	Recorder *httptest.ResponseRecorder
}

func (s *SessionManagerSuiteT) SetupTest() {
	support.Init(nil)
	s.Config = &support.ConfigT{}
	support.DeepClone(&s.Config, &support.Config)
	s.Config.HTTPSessionSecrets = [][]byte{[]byte("a401a91b016dcd4e6fcf4b96bf1ae283")}
	s.Recorder = httptest.NewRecorder()
}

func (s *SessionManagerSuiteT) TestSessionUnknownStore() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "unknown"
	s.Panics(func() { SessionManager(s.Config)(ctx) })
}

func (s *SessionManagerSuiteT) TestSessionCookieStore() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "cookie"
	SessionManager(s.Config)(ctx)

	session := DefaultSession(ctx)
	session.Options(ginsessions.Options{
		MaxAge: 368400,
	})
	sessionCookie, _ := ctx.Cookie(s.Config.HTTPSessionName)
	s.NotNil(sessionCookie)
	testSessionOps(s, session)
}

func (s *SessionManagerSuiteT) TestSessionRedisStore() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "redis"
	SessionManager(s.Config)(ctx)

	session := DefaultSession(ctx)
	testSessionOps(s, session)
}

func (s *SessionManagerSuiteT) TestSessionRedisStoreWrongAddr() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "redis"
	s.Config.HTTPSessionRedisAddr = "0.0.0.0:1234"
	s.Panics(func() { SessionManager(s.Config)(ctx) })
}

func (s *SessionManagerSuiteT) TestSessionRedisStoreWrongAuth() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "redis"
	s.Config.HTTPSessionRedisAuth = "authme"
	s.Panics(func() { SessionManager(s.Config)(ctx) })
}

func (s *SessionManagerSuiteT) TestSessionRedisStoreInvalidDb() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{}
	s.Config.HTTPSessionProvider = "redis"
	s.Config.HTTPSessionRedisDb = "-1"
	s.Panics(func() { SessionManager(s.Config)(ctx) })
}

func testSessionOps(s *SessionManagerSuiteT, session Session) {
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

func TestSessionManager(t *testing.T) {
	test.Run(t, new(SessionManagerSuiteT))
}
