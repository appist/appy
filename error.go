package appy

import "errors"

var (
	// ErrMissingMasterKey indicates the master key is not provided.
	ErrMissingMasterKey = errors.New("master key is missing")

	// ErrNoEmbeddedAssets indicates the embedded asset is missing.
	ErrNoEmbeddedAssets = errors.New("embedded asset is missing")

	// ErrReadMasterKeyFile indicates there is a problem reading master key file.
	ErrReadMasterKeyFile = errors.New("failed to read master key file in config path")
)
