package appy

import "errors"

var (
	// ErrNoMasterKey indicates the master key is not provided.
	ErrNoMasterKey = errors.New("master key is not provided")

	// ErrReadMasterKeyFile indicates there is a problem reading master key file.
	ErrReadMasterKeyFile = errors.New("failed to read master key file")

	// ErrNoConfigInAssets indicates the config is missing in the assets.
	ErrNoConfigInAssets = errors.New("missing config in the assets")
)
