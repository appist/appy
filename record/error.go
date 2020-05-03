package record

import "errors"

var (
	// ErrModelEmptyQueryBuilder indicates the model's query builder is empty. To fix
	// the error, simply call:
	//
	// - All
	// - Count
	// - Create
	// - Find
	// - Update
	ErrModelEmptyQueryBuilder = errors.New("model's query builder is empty")

	// ErrModelMissingMasterDB indicates the model is missing master database.
	ErrModelMissingMasterDB = errors.New("model is missing master database")

	// ErrModelMissingReplicaDB indicates the model is missing replica database.
	ErrModelMissingReplicaDB = errors.New("model is missing replica database")
)
