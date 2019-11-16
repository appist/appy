package test

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type (
	// Suite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
	Suite = suite.Suite
)

var (
	// CreateHTTPContext returns a fresh router w/ context for testing purposes.
	CreateHTTPContext = gin.CreateTestContext

	// RunSuite takes a testing suite and runs all of the tests attached to it.
	RunSuite = suite.Run
)
