package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AssertionT = assert.Assertions
type RequireT = require.Assertions
type SuiteT = suite.Suite
type AfterT = suite.AfterTest
type BeforeT = suite.BeforeTest
type SetupAllT = suite.SetupAllSuite
type SetupEachT = suite.SetupTestSuite
type TearDownAllT = suite.TearDownAllSuite
type TearDownEachT = suite.TearDownTestSuite
type TestingSuitT = suite.TestingSuite

var Run = suite.Run

func NewAssert(t *testing.T) *AssertionT {
	return assert.New(t)
}

func NewRequire(t *testing.T) *RequireT {
	return require.New(t)
}
