package middleware

import (
	"{{.Project.Name}}/pkg/app"
	"bufio"
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy"
)

type ExampleSuite struct {
	appy.TestSuite
	buffer   *bytes.Buffer
	logger   *appy.Logger
	recorder *httptest.ResponseRecorder
	writer   *bufio.Writer
}

func (s *ExampleSuite) SetupTest() {
	s.logger, s.buffer, s.writer = appy.NewFakeLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *ExampleSuite) TearDownTest() {
}

func (s *ExampleSuite) TestExample() {
	oldLogger := app.Logger
	app.Logger = s.logger
	defer func() { app.Logger = oldLogger }()

	c, _ := appy.NewTestContext(s.recorder)
	Example()(c)
	s.writer.Flush()

	s.Contains(s.buffer.String(), "middleware example logging")
}

func TestExampleSuite(t *testing.T) {
	appy.RunTestSuite(t, new(ExampleSuite))
}
