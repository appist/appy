package support

import (
	"reflect"
	"testing"

	"github.com/appist/appy/test"
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
	s.Equal(true, ok)
}

func (s *LoggerSuite) TestNewFakeLogger() {
	logger, buf, writer := NewFakeLogger()
	logger.Info("test")
	writer.Flush()
	s.NotNil(logger)
	s.Contains(buf.String(), "test")
}

func (s *LoggerSuite) TestNewLoggerConfigWithReleaseBuild() {
	oldBuild := Build
	Build = ReleaseBuild
	defer func() {
		Build = oldBuild
	}()

	config := newLoggerConfig()
	s.NotNil(config.EncoderConfig.EncodeTime)
}

func TestLoggerSuite(t *testing.T) {
	test.RunSuite(t, new(LoggerSuite))
}
