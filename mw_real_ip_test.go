package appy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy"
)

type RealIPSuite struct {
	appy.TestSuite
	recorder *httptest.ResponseRecorder
}

func (s *RealIPSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *RealIPSuite) TearDownTest() {
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXForwardedForIsSet() {
	ctx, _ := appy.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Forwarded-For", "localhost")
	appy.RealIP()(ctx)
	s.Equal("localhost", ctx.Request.RemoteAddr)
}

func (s *RealIPSuite) TestRemoteAddressNotNilIfXRealIPIsSet() {
	ctx, _ := appy.CreateTestContext(s.recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Real-IP", "localhost")
	appy.RealIP()(ctx)
	s.Equal("localhost", ctx.Request.RemoteAddr)
}

func TestRealIP(t *testing.T) {
	appy.RunTestSuite(t, new(RealIPSuite))
}
