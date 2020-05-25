package middleware

import (
	"bufio"
	"bytes"
	"{{.projectName}}/pkg/app"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type ExampleSuite struct {
	test.Suite
	buffer   *bytes.Buffer
	logger   *support.Logger
	recorder *httptest.ResponseRecorder
	writer   *bufio.Writer
}

func (s *ExampleSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
	s.recorder = httptest.NewRecorder()
}

func (s *ExampleSuite) TearDownTest() {
}

func (s *ExampleSuite) TestExample() {
	oldLogger := app.Logger
	app.Logger = s.logger
	defer func() { app.Logger = oldLogger }()

	c, _ := pack.NewTestContext(s.recorder)
	Example()(c)
	s.writer.Flush()

	s.Contains(s.buffer.String(), "middleware example logging")
}

func TestExampleSuite(t *testing.T) {
	test.Run(t, new(ExampleSuite))
}
