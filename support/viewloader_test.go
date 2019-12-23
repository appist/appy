package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type ViewLoaderSuite struct {
	test.Suite
}

func (s *ViewLoaderSuite) SetupTest() {
}

func (s *ViewLoaderSuite) TearDownTest() {
}

func (s *ViewLoaderSuite) TestNewViewEngine() {
	assets := NewAssets(nil, "", nil)
	viewLoader := NewViewLoader(assets)
	s.NotNil(viewLoader)
	s.NotNil(viewLoader.assets)
}

func TestViewLoaderSuite(t *testing.T) {
	test.RunSuite(t, new(ViewLoaderSuite))
}
