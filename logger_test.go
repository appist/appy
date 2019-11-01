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
	l := appy.NewLogger(appy.DebugBuild)
	_, ok := reflect.TypeOf(l).MethodByName("Desugar")
	s.Equal(appy.DebugBuild, l.Build())
	s.Equal(true, l.DbLogging())
	s.Equal(true, ok)

	l = appy.NewLogger(appy.ReleaseBuild)
	s.Equal(appy.ReleaseBuild, l.Build())
	s.Equal(true, l.DbLogging())
}

func TestLoggerSuite(t *testing.T) {
	appy.RunTestSuite(t, new(LoggerSuite))
}
