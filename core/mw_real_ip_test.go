package core

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
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Forwarded-For", "0.0.0.0")
	RealIP()(ctx)
	s.Equal("0.0.0.0", ctx.Request.RemoteAddr)
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXRealIPIsSet() {
	ctx, _ := test.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Real-IP", "0.0.0.0")
	RealIP()(ctx)
	s.Equal("0.0.0.0", ctx.Request.RemoteAddr)
}

func TestRealIP(t *testing.T) {
	test.Run(t, new(RealIPSuite))
}
