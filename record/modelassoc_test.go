package record

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type (
	modelAssocSuite struct {
		test.Suite
		buffer    *bytes.Buffer
		dbManager *Engine
		i18n      *support.I18n
		logger    *support.Logger
		writer    *bufio.Writer
	}
)

func (s *modelAssocSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
	asset := support.NewAsset(nil, "testdata")
	config := support.NewConfig(asset, s.logger)
	s.i18n = support.NewI18n(asset, config, s.logger)
}

func (s *modelAssocSuite) TearDownTest() {
	for _, database := range s.dbManager.Databases() {
		s.Nil(database.Close())
	}
}

func (s *modelAssocSuite) model(v interface{}, opts ...ModelOption) Modeler {
	return NewModel(s.dbManager, v, opts...)
}

func TestModelAssocSuite(t *testing.T) {
	test.Run(t, new(modelAssocSuite))
}
