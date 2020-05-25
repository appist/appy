package middleware

import (
	"bufio"
	"bytes"
	"context"
	"{{.projectName}}/pkg/app"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/appist/appy/worker"
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

	ctx := context.Background()
	job := worker.NewJob("test", nil)

	mockedHandler := worker.NewMockedHandler()
	mockedHandler.On("ProcessTask", ctx, job).Return(nil)
	err := Example(mockedHandler).ProcessTask(ctx, job)
	s.writer.Flush()

	s.Nil(err)
	s.Contains(s.buffer.String(), "middleware example logging")
	mockedHandler.AssertExpectations(s.T())
}

func TestExampleSuite(t *testing.T) {
	test.Run(t, new(ExampleSuite))
}
