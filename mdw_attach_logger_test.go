package appy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AttachLoggerSuite struct {
	TestSuite
	logger   *Logger
	buffer   *bytes.Buffer
	recorder *httptest.ResponseRecorder
}

func (s *AttachLoggerSuite) SetupTest() {
	s.logger, s.buffer, _ = NewFakeLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *AttachLoggerSuite) TestAttachLogger() {
	c, _ := NewTestContext(s.recorder)
	c.Request = &http.Request{
		Header: map[string][]string{},
	}
	s.Nil(c.Logger())
	AttachLogger(s.logger)(c)
	s.NotNil(c.Logger())
	c.Logger().Info("testing")
	s.Contains("testing", s.buffer.String())
}

func TestAttachLogger(t *testing.T) {
	RunTestSuite(t, new(AttachLoggerSuite))
}
