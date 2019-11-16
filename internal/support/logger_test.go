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
	logger := NewLogger()
	_, ok := reflect.TypeOf(logger).MethodByName("Desugar")
	s.Equal(true, logger.DbLogging())
	s.Equal(true, ok)

	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()
	logger = NewLogger()
	s.Equal(true, logger.DbLogging())
}

func (s *LoggerSuite) TestNewFakeLogger() {
	logger, buf, writer := NewFakeLogger()
	logger.Info("test")
	writer.Flush()
	s.NotNil(logger)
	s.Contains(buf.String(), "test")
}

func (s *LoggerSuite) TestSetDbLogging() {
	logger := NewLogger()
	s.Equal(true, logger.DbLogging())
	logger.SetDbLogging(false)
	s.Equal(false, logger.DbLogging())
}

func TestLoggerSuite(t *testing.T) {
	test.RunSuite(t, new(LoggerSuite))
}
