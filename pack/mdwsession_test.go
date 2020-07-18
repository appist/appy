package pack

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	gorsessions "github.com/gorilla/sessions"
)

type mdwSessionSuite struct {
	test.Suite
	asset    *support.Asset
	config   *support.Config
	logger   *support.Logger
	recorder *httptest.ResponseRecorder
}

func testOps(s *mdwSessionSuite, session Sessioner) {
	s.Nil(session.Save())
	s.NotNil(session.Values())

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

func (s *mdwSessionSuite) SetupTest() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_REDIS_ADDR", "0.0.0.0:16379")

	s.logger = support.NewLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.recorder = httptest.NewRecorder()
}

func (s *mdwSessionSuite) TearDownTest() {
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwSessionSuite) TestSessionUnknownStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}

	s.config.HTTPSessionProvider = "unknown"
	s.Panics(func() { mdwSession(s.config)(c) })
}

func (s *mdwSessionSuite) TestSessionCookieStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "cookie"
	mdwSession(s.config)(c)

	session := c.Session()
	session.Options(SessionOptions{
		MaxAge: 368400,
	})
	sessionCookie, _ := c.Cookie(s.config.HTTPSessionCookieName)

	s.NotNil(sessionCookie)
	testOps(s, session)
}

func (s *mdwSessionSuite) TestSessionRedisStore() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	mdwSession(s.config)(c)
	session := c.Session()

	testOps(s, session)
}

func (s *mdwSessionSuite) TestSessionRedisStoreWrongAddr() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisAddr = "localhost:1234"

	s.Panics(func() { mdwSession(s.config)(ctx) })
}

func (s *mdwSessionSuite) TestSessionRedisStoreWrongAuth() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	s.config.HTTPSessionRedisPassword = "authme"

	s.Panics(func() { mdwSession(s.config)(c) })
}

func (s *mdwSessionSuite) TestNonExistentSession() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"

	s.Nil(c.Session())
}

func (s *mdwSessionSuite) TestCustomSessionKey() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}
	s.config.HTTPSessionProvider = "redis"
	mdwSession(s.config)(c)

	session := c.Session()
	session.SetKeyPrefix("mysession:")
	testOps(s, session)

	s.Contains(session.Key(), "mysession:")
}

type fakeSessionStore struct {
	keyPrefix string
}

func (fss *fakeSessionStore) New(r *http.Request, name string) (*gorsessions.Session, error) {
	return gorsessions.GetRegistry(r).Get(fss, name)
}
func (fss *fakeSessionStore) Get(r *http.Request, name string) (*gorsessions.Session, error) {
	return nil, errors.New("not implemented")
}
func (fss *fakeSessionStore) KeyPrefix() string      { return fss.keyPrefix }
func (fss *fakeSessionStore) SetKeyPrefix(p string)  { fss.keyPrefix = p }
func (fss *fakeSessionStore) Options(SessionOptions) {}
func (fss *fakeSessionStore) Save(r *http.Request, w http.ResponseWriter, session *gorsessions.Session) error {
	return nil
}

func (s *mdwSessionSuite) TestSessionStoreUnableToReturnSession() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{}

	store := &fakeSessionStore{}
	session := &Session{
		store: store,
	}
	s.Nil(session.Values())
}

func TestMdwSessionSuite(t *testing.T) {
	test.Run(t, new(mdwSessionSuite))
}
