package mocks

import record "github.com/appist/appy/record"

// NewModel initializes a test model that is useful for testing purpose.
func NewModel() (func(dest interface{}, opts ...record.ModelOption) record.Modeler, *Model) {
	mm := &Model{}

	return func(dest interface{}, opts ...record.ModelOption) record.Modeler {
		return mm
	}, mm
}
