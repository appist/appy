package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type ParserSuite struct {
	test.Suite
}

func (s *ParserSuite) SetupTest() {
}

func (s *ParserSuite) TearDownTest() {
}

func (s *ParserSuite) TestParseEnvWithSupportedTypes() {
	type testConfig struct {
		Admins  map[string]string `env:"TEST_ADMINS" envDefault:"user1:pass1,user2:pass2"`
		Hosts   []string          `env:"TEST_HOSTS" envDefault:"0.0.0.0,1.1.1.1"`
		Secret  []byte            `env:"TEST_SECRET" envDefault:"hello"`
		Secrets [][]byte          `env:"TEST_SECRETS" envDefault:"hello,world"`
	}

	c := &testConfig{}
	ParseEnv(c)
	s.Equal(map[string]string{"user1": "pass1", "user2": "pass2"}, c.Admins)
	s.Equal([]string{"0.0.0.0", "1.1.1.1"}, c.Hosts)
	s.Equal([]byte("hello"), c.Secret)
	s.Equal([][]byte{[]byte("hello"), []byte("world")}, c.Secrets)
}

func (s *ParserSuite) TestParseEnvWithUnsupportedTypes() {
	type testConfig struct {
		Users map[string]int `env:"TEST_USERS" envDefault:"user1:1,user2:2"`
	}

	err := ParseEnv(&testConfig{})
	s.NotNil(err)
}

func (s *ParserSuite) TestParseEnvWithInvalidFormat() {
	type testConfig struct {
		Users map[string]string `env:"TEST_USERS" envDefault:"user1"`
	}

	c := &testConfig{}
	ParseEnv(c)
	s.Equal(map[string]string{}, c.Users)
}

func TestParserSuite(t *testing.T) {
	test.RunSuite(t, new(ParserSuite))
}
