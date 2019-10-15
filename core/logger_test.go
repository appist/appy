package core

import (
	"reflect"
	"testing"

	"github.com/appist/appy/test"
)

type LoggerSuite struct {
	test.Suite
}

func (s *LoggerSuite) SetupTest() {
	Build = "debug"
}

func (s *LoggerSuite) TearDownTest() {
}

func (s *LoggerSuite) TestNewLoggerConfig() {
	c := newLoggerConfig()
	s.Equal(true, c.Development)

	Build = "release"
	c = newLoggerConfig()
	s.Equal(false, c.Development)
}

func (s *LoggerSuite) TestNewLogger() {
	l, _ := newLogger(newLoggerConfig())
	_, ok := reflect.TypeOf(l).MethodByName("Desugar")
	s.Equal(true, ok)

	c := newLoggerConfig()
	c.Encoding = "test"
	_, err := newLogger(c)
	s.NotNil(err)
}

func TestLogger(t *testing.T) {
	test.Run(t, new(LoggerSuite))
}
