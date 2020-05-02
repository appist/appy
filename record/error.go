package record

import "errors"

var (
	// ErrMissingModelDB indidates the model is missing masters/replicas database.
	ErrMissingModelDB = errors.New("model is missing masters/replicas database")
)
