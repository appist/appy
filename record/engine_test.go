package record

import (
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type engineSuite struct {
	test.Suite
	logger *support.Logger
}

func (s *engineSuite) SetupTest() {
	s.logger, _, _ = support.NewTestLogger()
}

func (s *engineSuite) TestDBManagerWithBadDBConfig() {
	os.Setenv("DB_URI_PRIMARY", "0.0.0.0:13306/appy")
	defer os.Unsetenv("DB_URI_PRIMARY")

	engine := NewEngine(s.logger, nil)
	s.Equal(1, len(engine.Errors()))
}

func (s *engineSuite) TestDBManagerWithValidDBConfig() {
	os.Setenv("DB_URI_PRIMARY", "mysql://root:whatever@0.0.0.0:13306/appy")
	os.Setenv("DB_URI_PRIMARY_REPLICA", "mysql://root:whatever@0.0.0.0:13307/appy")
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_URI_PRIMARY_REPLICA")
	}()

	engine := NewEngine(s.logger, nil)
	s.Equal(0, len(engine.Errors()))
	s.Equal(2, len(engine.Databases()))
	s.Nil(engine.DB("foobar"))
	s.NotNil(engine.DB("primary"))
	s.NotNil(engine.DB("primaryReplica"))
	s.Contains(engine.Info(), "* DBs:")
	s.Contains(engine.Info(), "primary")
	s.Contains(engine.Info(), "primaryReplica")
}

func TestDBManagerSuite(t *testing.T) {
	test.Run(t, new(engineSuite))
}
