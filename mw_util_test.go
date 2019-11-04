package appy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type UtilSuite struct {
	TestSuite
}

func (s *UtilSuite) SetupTest() {
}

func (s *UtilSuite) TearDownTest() {
}

func (s *UtilSuite) TestIsAPIOnly() {
	recorder := httptest.NewRecorder()

	ctx, _ := CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "1")
	s.Equal(true, IsAPIOnly(ctx))

	ctx, _ = CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "0")
	s.Equal(false, IsAPIOnly(ctx))
}

func TestUtil(t *testing.T) {
	RunTestSuite(t, new(UtilSuite))
}
