package record

import "errors"

var (
	ErrMissingModelDB = errors.New("model missing master/replica database")
)
