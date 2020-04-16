package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type (
	// Assert provides assertion methods around the *testing.T interface.
	Assert struct {
		assert.Assertions
	}

	// Mock is the workhorse used to track activity on another object.
	Mock struct {
		mock.Mock
	}

	// Suite is a basic testing suite with methods for storing and retrieving
	// the current *testing.T context.
	Suite struct {
		suite.Suite
	}
)

var (
	// Run takes a testing suite and runs all of the tests attached to it.
	Run = suite.Run
)
