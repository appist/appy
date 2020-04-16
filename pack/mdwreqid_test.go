package pack

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/test"
)

type mdwReqIDSuite struct {
	test.Suite
	recorder *httptest.ResponseRecorder
}

func (s *mdwReqIDSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func (s *mdwReqIDSuite) TestMdwReqID() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	s.Empty(c.RequestID())

	mdwReqID()(c)
	s.NotEmpty(c.RequestID())
}

func TestMdwReqIDSuite(t *testing.T) {
	test.Run(t, new(mdwReqIDSuite))
}
