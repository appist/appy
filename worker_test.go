package appy

import "testing"

type WorkerSuite struct {
	TestSuite
}

func (s *WorkerSuite) SetupTest() {
}

func (s *WorkerSuite) TearDownTest() {
}

func (s *WorkerSuite) TestNewWorker() {
}

func TestWorkerSuite(t *testing.T) {
	RunTestSuite(t, new(WorkerSuite))
}
