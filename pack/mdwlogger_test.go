package pack

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mdwLoggerSuite struct {
	test.Suite
	logger   *support.Logger
	buffer   *bytes.Buffer
	recorder *httptest.ResponseRecorder
}

func (s *mdwLoggerSuite) SetupTest() {
	s.logger, s.buffer, _ = support.NewTestLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *mdwLoggerSuite) TestMdwLogger() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	s.Nil(c.Logger())

	mdwLogger(s.logger)(c)
	s.NotNil(c.Logger())

	c.Logger().Info("testing")
	s.Contains("testing", s.buffer.String())
}

func TestMdwLoggerSuite(t *testing.T) {
	test.Run(t, new(mdwLoggerSuite))
}
