package appy

import "testing"

type WorkerLoggerSuite struct {
	TestSuite
}

func (s *WorkerLoggerSuite) SetupTest() {
}

func (s *WorkerLoggerSuite) TearDownTest() {
}

func (s *WorkerLoggerSuite) TestNewWorkerLogger() {
}

func TestWorkerLoggerSuite(t *testing.T) {
	RunTestSuite(t, new(WorkerLoggerSuite))
}
