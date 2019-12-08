package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type StringSuite struct {
	test.Suite
}

func (s *StringSuite) SetupTest() {
}

func (s *StringSuite) TearDownTest() {
}

func (s *StringSuite) TestIsCamelCase() {
	tt := [][]interface{}{
		{"fooBar", true},
		{"fooBar1", true},
		{"foo1Bar", true},
		{"", false},
		{"1FooBar", false},
		{"1fooBar", false},
		{"foo@Bar", false},
		{"foo_bar", false},
		{"FOO_BAR", false},
		{"FooBar", false},
		{"Foo@Bar", false},
	}

	for _, t := range tt {
		s.Equal(t[1], IsCamelCase(t[0].(string)))
	}
}

func (s *StringSuite) TestIsChainCase() {
	tt := [][]interface{}{
		{"foo-bar", true},
		{"foo1-bar2", true},
		{"foo-bar-1", true},
		{"foo-bar-1", true},
		{"111-foo-bar", true},
		{"foobar", true},
		{"foobar1", true},
		{"foo1bar", true},
		{"1foobar", true},
		{"", false},
		{"FOO-BAR", false},
		{"fooBar", false},
		{"FooBar", false},
		{"FOOBAR", false},
		{"foo-@bar", false},
		{"foo_bar", false},
		{"テスト", false},
		{"テスト-テスト", false},
	}

	for _, t := range tt {
		s.Equal(t[1], IsChainCase(t[0].(string)))
	}
}

func (s *StringSuite) TestIsFlatCase() {
	tt := [][]interface{}{
		{"foobar", true},
		{"foo1bar", true},

		{"", false},
		{"1foobar", false},
		{"foo@bar", false},
		{"foo_bar", false},
		{"FOO_BAR", false},
		{"FooBar", false},
		{"fooBar", false},
		{"foo_bar", false},
	}

	for _, t := range tt {
		s.Equal(t[1], IsFlatCase(t[0].(string)))
	}
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

func (s *StringSuite) TestIsSnakeCase() {
	tt := [][]interface{}{
		{"foo_bar", true},
		{"foo1_bar2", true},
		{"foo_bar_1", true},
		{"foo_bar_1", true},
		{"111_foo_bar", true},
		{"foobar", true},
		{"foobar1", true},
		{"foo1bar", true},
		{"111foobar", true},
		{"", false},
		{"FOO_BAR", false},
		{"fooBar", false},
		{"FooBar", false},
		{"FOOBAR", false},
		{"foo_@bar", false},
		{"foo-bar", false},
		{"テスト", false},
		{"テスト_テスト", false},
	}

	for _, t := range tt {
		s.Equal(t[1], IsSnakeCase(t[0].(string)))
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

func (s *StringSuite) TestToChainCase() {
	tt := [][]string{
		{"foo_bar", "foo-bar"},
		{"foo-bar", "foo-bar"},
		{"foo-bar_baz", "foo-bar-baz"},
		{"foo--bar__baz", "foo-bar-baz"},
		{"fooBar", "foo-bar"},
		{"FooBar", "foo-bar"},
		{"foo bar", "foo-bar"},
		{"   foo   bar   ", "foo-bar"},
		{"fooBar111", "foo-bar-111"},
		{"111FooBar", "111-foo-bar"},
		{"foo-111-Bar", "foo-111-bar"},
		{"", ""},
	}

	for _, t := range tt {
		s.Equal(t[1], ToChainCase(t[0]))
	}
}

func (s *StringSuite) TestToFlatCase() {
	tt := [][]string{
		{"foo_bar", "foobar"},
		{"foo-bar", "foobar"},
		{"foo-bar_baz", "foobarbaz"},
		{"foo--bar__baz", "foobarbaz"},
		{"fooBar", "foobar"},
		{"FooBar", "foobar"},
		{"foo bar", "foobar"},
		{"   foo   bar   ", "foobar"},
		{"fooBar111", "foobar111"},
		{"111FooBar", "111foobar"},
		{"foo-111-Bar", "foo111bar"},
		{"", ""},
	}

	for _, t := range tt {
		s.Equal(t[1], ToFlatCase(t[0]))
	}
}

func (s *StringSuite) TestToPascalCase() {
	tt := [][]string{
		{"foo_bar", "FooBar"},
		{"foo-bar", "FooBar"},
		{"foo-bar_baz", "FooBarBaz"},
		{"foo--bar__baz", "FooBarBaz"},
		{"fooBar", "FooBar"},
		{"FooBar", "FooBar"},
		{"foo bar", "FooBar"},
		{"   foo   bar   ", "FooBar"},
		{"fooBar111", "FooBar111"},
		{"111FooBar", "111FooBar"},
		{"foo-111-Bar", "Foo111Bar"},
		{"", ""},
	}

	for _, t := range tt {
		s.Equal(t[1], ToPascalCase(t[0]))
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

func (s *StringSuite) TestIsFirstRuneLower() {
	str := ""
	s.Equal(false, isFirstRuneLower(str))

	str = "test"
	s.Equal(true, isFirstRuneLower(str))
}

func (s *StringSuite) TestIsFirstRuneUpper() {
	str := ""
	s.Equal(false, isFirstRuneUpper(str))

	str = "Test"
	s.Equal(true, isFirstRuneUpper(str))
}

func (s *StringSuite) TestRuneAt() {
	str := ""
	s.Equal(int32(0), runeAt(str, 0))

	str = "Test"
	s.Equal(int32(84), runeAt(str, 0))
}

func TestStringSuite(t *testing.T) {
	test.RunSuite(t, new(StringSuite))
}
