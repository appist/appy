package appy

import "errors"

var (
	ErrNoMasterKey       = errors.New("master key is not provided")
	ErrReadMasterKeyFile = errors.New("failed to read master key file")
	ErrNoConfigInAssets  = errors.New("missing config in the assets")
)
