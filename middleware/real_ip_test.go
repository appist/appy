package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type RealIPSuiteT struct {
	test.SuiteT
	Recorder *httptest.ResponseRecorder
}

func (s *RealIPSuiteT) SetupTest() {
	s.Recorder = httptest.NewRecorder()
}

func (s *RealIPSuiteT) TestRemoteAddressNotNilIfXForwardedForIsSet() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Forwarded-For", "0.0.0.0")
	RealIP()(ctx)
	s.Equal("0.0.0.0", ctx.Request.RemoteAddr)
}

func (s *RealIPSuiteT) TestRemoteAddressNotNilIfXRealIPIsSet() {
	ctx, _ := test.CreateContext(s.Recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Set("X-Real-IP", "0.0.0.0")
	RealIP()(ctx)
	s.Equal("0.0.0.0", ctx.Request.RemoteAddr)
}

func TestRealIP(t *testing.T) {
	test.Run(t, new(RealIPSuiteT))
}
