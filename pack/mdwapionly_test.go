package pack

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type mdwAPIOnlySuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *mdwAPIOnlySuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *mdwAPIOnlySuite) TestAPIOnly() {
	{
		c, _ := NewTestContext(s.recorder)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Writer.Header().Add("Set-Cookie", "test")
		mdwAPIOnly()(c)

		s.Equal("test", c.Writer.Header().Get("Set-Cookie"))
	}

	{
		c, _ := NewTestContext(s.recorder)
		c.Request = &http.Request{
			Header: map[string][]string{},
		}
		c.Request.Header.Add("X-API-Only", "true")
		c.Writer.Header().Add("Set-Cookie", "test")
		mdwAPIOnly()(c)

		s.Equal("", c.Writer.Header().Get("Set-Cookie"))
	}
}

func TestMdwAPIOnlySuite(t *testing.T) {
	test.Run(t, new(mdwAPIOnlySuite))
}
