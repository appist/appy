package appy

import (
	"bufio"
	"bytes"
	"os"
	"testing"
)

type WorkerLoggerSuite struct {
	TestSuite
	logger *WorkerLogger
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func (s *WorkerLoggerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	baseLogger, buffer, writer := NewFakeLogger()
	asset := NewAsset(nil, map[string]string{
		"docker": "testdata/server/.docker",
		"config": "testdata/server/configs",
		"locale": "testdata/server/pkg/locales",
		"view":   "testdata/server/pkg/views",
		"web":    "testdata/server/web",
	}, "")
	config := NewConfig(asset, baseLogger, &Support{})
	worker := NewWorker(asset, config, baseLogger)

	s.buffer = buffer
	s.writer = writer
	s.logger = &WorkerLogger{
		baseLogger,
		worker,
	}
}

func (s *WorkerLoggerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *WorkerLoggerSuite) TestDebug() {
	s.logger.Debug("test %s", "foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "DEBUG")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *WorkerLoggerSuite) TestInfo() {
	s.logger.Info("test %s", "foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *WorkerLoggerSuite) TestWorkerInfo() {
	skipMessages := []string{
		"Bye!",
		"heartbeat done",
		"i am shutting down...",
		"Waiting for all workers to finish...",
		"All workers have finished",
		"Send signal TSTP to",
		"Starting processing",
	}

	for _, skipMessage := range skipMessages {
		s.logger.Info(skipMessage)
		s.writer.Flush()
		s.Equal("", s.buffer.String())
	}

	s.logger.Info("Starting graceful shutdown")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "* Gracefully shutting down the worker...")

	s.logger.Info("Send signal TERM to")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "* appy 0.1.0 (go1.14), build: debug, environment: development, config: testdata/server/configs/.env.development")
	s.Contains(s.buffer.String(), "* Concurrency: 25, queues: default=10")
	s.Contains(s.buffer.String(), "* Worker is now ready to process jobs...")
}

func (s *WorkerLoggerSuite) TestWarn() {
	s.logger.Warn("test %s", "foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "WARN")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *WorkerLoggerSuite) TestError() {
	s.logger.Error("test %s", "foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "ERROR")
	s.Contains(s.buffer.String(), "test foo")
}

func TestWorkerLoggerSuite(t *testing.T) {
	RunTestSuite(t, new(WorkerLoggerSuite))
}
