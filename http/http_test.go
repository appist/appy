package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type HTTPSuite struct {
	test.Suite
}

func (s *HTTPSuite) SetupTest() {
}

func (s *HTTPSuite) TearDownTest() {
}

func (s *HTTPSuite) TestIsAPIOnly() {
	recorder := httptest.NewRecorder()

	c, _ := NewTestContext(recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Add("X-API-Only", "1")
	s.Equal(true, IsAPIOnly(c))

	c, _ = NewTestContext(recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	c.Request.Header.Add("X-API-Only", "0")
	s.Equal(false, IsAPIOnly(c))
}

func TestHTTPSuite(t *testing.T) {
	test.RunSuite(t, new(HTTPSuite))
}
