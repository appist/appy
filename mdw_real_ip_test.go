package appy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type RealIPSuite struct {
	TestSuite
	recorder *httptest.ResponseRecorder
}

func (s *RealIPSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXForwardedForIsSet() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("X-Forwarded-For", "localhost")
	RealIP()(c)
	s.Equal("localhost", c.Request.RemoteAddr)
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXRealIPIsSet() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Set("X-Real-IP", "localhost")
	RealIP()(c)
	s.Equal("localhost", c.Request.RemoteAddr)
}

func TestRealIP(t *testing.T) {
	RunTestSuite(t, new(RealIPSuite))
}
