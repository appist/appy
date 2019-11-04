package appy

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerSuite struct {
	TestSuite
}

func newMockedLogger() (*Logger, *bytes.Buffer, *bufio.Writer) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	return &Logger{
		SugaredLogger: zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(writer),
				zapcore.DebugLevel,
			),
		).Sugar(),
	}, &buffer, writer
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

func (s *LoggerSuite) TestSetDbLogging() {
	logger := NewLogger(DebugBuild)
	s.Equal(true, logger.DbLogging())
	logger.SetDbLogging(false)
	s.Equal(false, logger.DbLogging())
}

func TestLoggerSuite(t *testing.T) {
	RunTestSuite(t, new(LoggerSuite))
}
