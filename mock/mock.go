package mock

import record "github.com/appist/appy/record"

// NewDB initializes a test DB that is useful for testing purpose.
func NewDB() (func(name string) record.DBer, *DB) {
	db := &DB{}

	return func(name string) record.DBer {
		return db
	}, db
}

// NewModel initializes a test model that is useful for testing purpose.
func NewModel() (func(dest interface{}, opts ...record.ModelOption) record.Modeler, *Model) {
	m := &Model{}

	return func(dest interface{}, opts ...record.ModelOption) record.Modeler {
		return m
	}, m
}
