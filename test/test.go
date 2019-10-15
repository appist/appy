package test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Assert provides assertion methods around the TestingT interface.
type Assert = assert.Assertions

// Suite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
type Suite = suite.Suite

// Run takes a testing suite and runs all of the tests attached to it.
var Run = suite.Run

// CreateTestContext returns a fresh router engine and context for testing purposes.
var CreateTestContext = gin.CreateTestContext

// NewAssert makes a new AssertionT object for the specified TestingT.
func NewAssert(t *testing.T) *Assert {
	return assert.New(t)
}
