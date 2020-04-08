package pack

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type mdwRealIPSuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *mdwRealIPSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *mdwRealIPSuite) TestRemoteAddressNotNilIfXForwardedForIsSet() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("X-Forwarded-For", "localhost")
	mdwRealIP()(c)

	s.Equal("localhost", c.Request.RemoteAddr)
}

func (s *mdwRealIPSuite) TestRemoteAddressNotNilIfXRealIPIsSet() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("X-Real-IP", "localhost")
	mdwRealIP()(c)

	s.Equal("localhost", c.Request.RemoteAddr)
}

func TestMdwRealIPSuite(t *testing.T) {
	test.Run(t, new(mdwRealIPSuite))
}
