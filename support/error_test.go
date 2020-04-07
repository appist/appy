package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type ErrorSuite struct {
	test.Suite
}

func (s *ErrorSuite) TestErrorMessage() {
	s.Equal("master key is missing", ErrMissingMasterKey.Error())
	s.Equal("embedded asset is missing", ErrNoEmbeddedAssets.Error())
	s.Equal("failed to read master key file in config path", ErrReadMasterKeyFile.Error())
}

func TestErrorSuite(t *testing.T) {
	test.Run(t, new(ErrorSuite))
}
