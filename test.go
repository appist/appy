package appy

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type (
	// TestSuite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
	TestSuite = suite.Suite
)

var (
	// CreateTestContext returns a fresh router w/ context for testing purposes.
	CreateTestContext = gin.CreateTestContext

	// RunTestSuite takes a testing suite and runs all of the tests attached to it.
	RunTestSuite = suite.Run
)
