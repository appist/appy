package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type ViewEngineSuite struct {
	test.Suite
}

func (s *ViewEngineSuite) SetupTest() {
}

func (s *ViewEngineSuite) TearDownTest() {
}

func (s *ViewEngineSuite) TestNewViewEngine() {
	assets := NewAssets(nil, "", nil)
	viewEngine := NewViewEngine(assets)
	s.NotNil(viewEngine)
	s.NotNil(viewEngine.Set)
}

func TestViewEngineSuite(t *testing.T) {
	test.RunSuite(t, new(ViewEngineSuite))
}
