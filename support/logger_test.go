package support

import (
	"reflect"
	"testing"

	"github.com/appist/appy/test"
)

func TestNewLoggerConfig(t *testing.T) {
	assert := test.NewAssert(t)
	config := NewLoggerConfig()
	assert.Equal(true, config.Development)
	if config.Development != true {
		t.Fail()
	}

	Build = "release"
	config = NewLoggerConfig()
	assert.Equal(false, config.Development)
}

func TestNewLogger(t *testing.T) {
	assert := test.NewAssert(t)
	logger, _ := NewLogger(NewLoggerConfig())
	_, ok := reflect.TypeOf(logger).MethodByName("Desugar")
	assert.Equal(true, ok)

	config := NewLoggerConfig()
	config.Encoding = "test"
	_, err := NewLogger(config)
	assert.NotNil(err)
}
