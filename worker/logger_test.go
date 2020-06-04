package worker

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type loggerSuite struct {
	test.Suite
	logger *logger
	buffer *bytes.Buffer
	writer *bufio.Writer
}

func (s *loggerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	baseLogger, buffer, writer := support.NewTestLogger()
	asset := support.NewAsset(nil, "")
	config := support.NewConfig(asset, baseLogger)
	dbManager := record.NewEngine(baseLogger, nil)
	worker := NewEngine(asset, config, dbManager, baseLogger)

	s.buffer = buffer
	s.writer = writer
	s.logger = &logger{
		baseLogger,
		worker,
	}
}

func (s *loggerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("DB_URI_PRIMARY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *loggerSuite) TestDebug() {
	s.logger.Debug("test foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "DEBUG")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *loggerSuite) TestInfo() {
	s.logger.Info("test foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "INFO")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *loggerSuite) TestWarn() {
	s.logger.Warn("test foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "WARN")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *loggerSuite) TestError() {
	s.logger.Error("test foo")
	s.writer.Flush()
	s.Contains(s.buffer.String(), "ERROR")
	s.Contains(s.buffer.String(), "test foo")
}

func (s *loggerSuite) TestWorkerInfo() {
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
	s.Contains(s.buffer.String(), fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: configs/.env.development", runtime.Version()))
	s.Contains(s.buffer.String(), "* DBs: primary")
	s.Contains(s.buffer.String(), "* Concurrency: 25, queues: default=10")
	s.Contains(s.buffer.String(), "* Worker is now ready to process jobs...")
}

func TestLoggerSuite(t *testing.T) {
	test.Run(t, new(loggerSuite))
}
