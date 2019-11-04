package appy_test

import (
	"reflect"
	"testing"

	"github.com/appist/appy"
)

type LoggerSuite struct {
	appy.TestSuite
}

func (s *LoggerSuite) SetupTest() {
}

func (s *LoggerSuite) TearDownTest() {
}

func (s *LoggerSuite) TestNewLogger() {
	logger := appy.NewLogger(appy.DebugBuild)
	_, ok := reflect.TypeOf(logger).MethodByName("Desugar")
	s.Equal(appy.DebugBuild, logger.Build())
	s.Equal(true, logger.DbLogging())
	s.Equal(true, ok)

	logger = appy.NewLogger(appy.ReleaseBuild)
	s.Equal(appy.ReleaseBuild, logger.Build())
	s.Equal(true, logger.DbLogging())
}

func (s *LoggerSuite) TestSetDbLogging() {
	logger := appy.NewLogger(appy.DebugBuild)
	s.Equal(true, logger.DbLogging())
	logger.SetDbLogging(false)
	s.Equal(false, logger.DbLogging())
}

func TestLoggerSuite(t *testing.T) {
	appy.RunTestSuite(t, new(LoggerSuite))
}
