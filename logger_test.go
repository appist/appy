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
	l := appy.NewLogger("debug")
	_, ok := reflect.TypeOf(l).MethodByName("Desugar")
	s.Equal("debug", l.Build())
	s.Equal(true, l.DbLogging())
	s.Equal(true, ok)

	l = appy.NewLogger("release")
	s.Equal("release", l.Build())
	s.Equal(true, l.DbLogging())
}

func TestLogger(t *testing.T) {
	appy.RunTestSuite(t, new(LoggerSuite))
}
