package support

import (
	"fmt"
	"testing"

	"github.com/appist/appy/test"
)

func TestInit(t *testing.T) {
	assert := test.NewAssert(t)

	t.Run("initiates Config and Logger singletons", func(t *testing.T) {
		fmt.Println(Build)
		Init(nil)
		assert.NotNil(Config)
		assert.NotNil(Logger)
	})
}
