package appy

import (
	"github.com/stretchr/testify/suite"
)

type (
	// TestSuite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
	TestSuite = suite.Suite
)

var (
	// RunTestSuite takes a testing suite and runs all of the tests attached to it.
	RunTestSuite = suite.Run
)
