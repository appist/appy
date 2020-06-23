package mock

import (
	"reflect"

	record "github.com/appist/appy/record"
	mock "github.com/stretchr/testify/mock"
)

type (
	// Mock is the workhorse used to track activity on another object.
	Mock struct {
		mock.Mock
	}

	// AnythingOfTypeArgument is a string that contains the type of an argument
	// for use when type checking.
	AnythingOfTypeArgument = mock.AnythingOfTypeArgument

	// IsTypeArgument is a struct that contains the type of an argument for use
	// when type checking. This is an alternative to AnythingOfType.
	IsTypeArgument = mock.IsTypeArgument
)

const (
	// Anything is used in Diff and Assert when the argument being tested
	// shouldn't be taken into consideration.
	Anything = mock.Anything
)

var (
	// AnythingOfType returns an AnythingOfTypeArgument object containing the
	// name of the type to check for.
	AnythingOfType = mock.AnythingOfType

	// IsType returns an IsTypeArgument object containing the type to check for.
	// You can provide a zero-value of the type to check.  This is an alternative
	// to AnythingOfType.
	IsType = mock.IsType
)

// NewDB initializes a test DB that is useful for testing purpose.
func NewDB() (func(name string) record.DBer, *DB) {
	db := &DB{}

	return func(name string) record.DBer {
		return db
	}, db
}

// NewModel initializes a test model that is useful for testing purpose.
func NewModel(mockedDest interface{}) (func(dest interface{}, opts ...record.ModelOption) record.Modeler, *Model) {
	m := &Model{}

	return func(dest interface{}, opts ...record.ModelOption) record.Modeler {
		if mockedDest != nil {
			val := reflect.ValueOf(dest)
			val.Elem().Set(reflect.ValueOf(mockedDest).Elem())
		}

		return m
	}, m
}
