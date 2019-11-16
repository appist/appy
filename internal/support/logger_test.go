package support

import (
	"reflect"
	"testing"

	"github.com/appist/appy/internal/test"
)

type LoggerSuite struct {
	test.Suite
}

func (s *LoggerSuite) SetupTest() {
}

func (s *LoggerSuite) TearDownTest() {
}

func (s *LoggerSuite) TestNewLogger() {
	logger := NewLogger(DebugBuild)
	_, ok := reflect.TypeOf(logger).MethodByName("Desugar")
	s.Equal(DebugBuild, logger.Build())
	s.Equal(true, logger.DbLogging())
	s.Equal(true, ok)

	logger = NewLogger(ReleaseBuild)
	s.Equal(ReleaseBuild, logger.Build())
	s.Equal(true, logger.DbLogging())
}

func (s *LoggerSuite) TestNewFakeLogger() {
	logger, buf, writer := NewFakeLogger()
	logger.Info("test")
	writer.Flush()
	s.NotNil(logger)
	s.Contains(buf.String(), "test")
}

func (s *LoggerSuite) TestBuild() {
	logger := NewLogger(DebugBuild)
	s.Equal(DebugBuild, logger.Build())

	logger = NewLogger(ReleaseBuild)
	s.Equal(ReleaseBuild, logger.Build())
}

func (s *LoggerSuite) TestSetDbLogging() {
	logger := NewLogger(DebugBuild)
	s.Equal(true, logger.DbLogging())
	logger.SetDbLogging(false)
	s.Equal(false, logger.DbLogging())
}

func TestLoggerSuite(t *testing.T) {
	test.RunSuite(t, new(LoggerSuite))
}
