package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type RealIPSuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *RealIPSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *RealIPSuite) TearDownTest() {
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXForwardedForIsSet() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Forwarded-For", "localhost")
	RealIP()(ctx)
	s.Equal("localhost", ctx.Request.RemoteAddr)
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXRealIPIsSet() {
	ctx, _ := NewTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Real-IP", "localhost")
	RealIP()(ctx)
	s.Equal("localhost", ctx.Request.RemoteAddr)
}

func TestRealIP(t *testing.T) {
	test.RunSuite(t, new(RealIPSuite))
}
