package appy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy"
)

type UtilSuite struct {
	appy.TestSuite
}

func (s *UtilSuite) SetupTest() {
}

func (s *UtilSuite) TearDownTest() {
}

func (s *UtilSuite) TestIsAPIOnly() {
	recorder := httptest.NewRecorder()

	ctx, _ := appy.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "1")
	s.Equal(true, appy.IsAPIOnly(ctx))

	ctx, _ = appy.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "0")
	s.Equal(false, appy.IsAPIOnly(ctx))
}

func TestUtil(t *testing.T) {
	appy.RunTestSuite(t, new(UtilSuite))
}
