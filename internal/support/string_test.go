package support

import (
	"testing"

	"github.com/appist/appy/internal/test"
)

type StringSuite struct {
	test.Suite
}

func (s *StringSuite) SetupTest() {
}

func (s *StringSuite) TearDownTest() {
}

func (s *StringSuite) TestIsPascalCase() {
	tt := [][]interface{}{
		{"FooBar", true},
		{"FooBar1", true},
		{"Foo1Bar", true},
		{"", false},
		{"1FooBar", false},
		{"1fooBar", false},
		{"Foo@Bar", false},
		{"foo_bar", false},
		{"FOO_BAR", false},
		{"fooBar", false},
		{"foo@Bar", false},
	}

	for _, t := range tt {
		s.Equal(t[1], IsPascalCase(t[0].(string)))
	}
}

func (s *StringSuite) TestToCamelCase() {
	tt := [][]string{
		{"foo_bar", "fooBar"},
		{"foo-bar", "fooBar"},
		{"foo-bar_baz", "fooBarBaz"},
		{"foo--bar__baz", "fooBarBaz"},
		{"fooBar", "fooBar"},
		{"FooBar", "fooBar"},
		{"foo bar", "fooBar"},
		{"   foo   bar   ", "fooBar"},
		{"fooBar111", "fooBar111"},
		{"111FooBar", "111FooBar"},
		{"foo-111-Bar", "foo111Bar"},
		{"", ""},
	}

	for _, t := range tt {
		s.Equal(t[1], ToCamelCase(t[0]))
	}
}

func (s *StringSuite) TestToSnakeCase() {
	tt := [][]string{
		{"foo_bar", "foo_bar"},
		{"foo-bar", "foo_bar"},
		{"foo-bar_baz", "foo_bar_baz"},
		{"foo--bar__baz", "foo_bar_baz"},
		{"fooBar", "foo_bar"},
		{"FooBar", "foo_bar"},
		{"foo bar", "foo_bar"},
		{"   foo   bar   ", "foo_bar"},
		{"fooBar111", "foo_bar_111"},
		{"111FooBar", "111_foo_bar"},
		{"foo-111-Bar", "foo_111_bar"},
		{"", ""},
	}

	for _, t := range tt {
		s.Equal(t[1], ToSnakeCase(t[0]))
	}
}

func TestStringSuite(t *testing.T) {
	test.RunSuite(t, new(StringSuite))
}
