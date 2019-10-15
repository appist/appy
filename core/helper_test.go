package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type HelperSuite struct {
	test.Suite
}

func (s *HelperSuite) SetupTest() {
	Build = "debug"
}

func (s *HelperSuite) TearDownTest() {
}

func (s *HelperSuite) TestIsAPIOnly() {
	recorder := httptest.NewRecorder()

	ctx, _ := test.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "1")
	s.Equal(true, IsAPIOnly(ctx))

	ctx, _ = test.CreateTestContext(recorder)
	ctx.Request = &http.Request{
		Header: map[string][]string{},
	}
	ctx.Request.Header.Add("X-API-Only", "0")
	s.Equal(false, IsAPIOnly(ctx))
}

func TestHelper(t *testing.T) {
	test.Run(t, new(HelperSuite))
}
