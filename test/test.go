package test

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type (
	// Suite is a basic testing suite with methods for storing and retrieving the current *testing.T context.
	Suite struct {
		suite.Suite
	}

	// Mock is the workhorse used to track activity on another object.
	Mock struct {
		mock.Mock
	}
)

var (
	// Run takes a testing suite and runs all of the tests attached to it.
	Run = suite.Run
)
