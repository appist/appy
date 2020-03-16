package appy_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy"
)

type SupportSuite struct {
	appy.TestSuite
	support *appy.Support
}

func (s *SupportSuite) SetupTest() {
	s.support = &appy.Support{}
}

func (s *SupportSuite) TestArrayContains() {
	tt := []struct {
		arr      interface{}
		val      interface{}
		expected bool
	}{
		{[]bool{}, true, false},
		{[]bool{true, true}, false, false},
		{[]bool{true, true}, true, true},
		{[]byte{}, 80, false},
		{[]byte{80, 81}, byte(82), false},
		{[]byte{80, 81}, byte(81), true},
		{[]complex64{}, 1 + 2i, false},
		{[]complex64{1 + 2i, 2 + 3i}, complex64(3 + 4i), false},
		{[]complex64{1 + 2i, 2 + 3i}, complex64(2 + 3i), true},
		{[]complex128{}, 1 + 2i, false},
		{[]complex128{1 + 2i, 2 + 3i}, complex128(3 + 4i), false},
		{[]complex128{1 + 2i, 2 + 3i}, complex128(2 + 3i), true},
		{[]float32{}, 0.1, false},
		{[]float32{0.1, 0.2}, float32(0.3), false},
		{[]float32{0.1, 0.2}, float32(0.1), true},
		{[]float64{}, 0.1, false},
		{[]float64{0.1, 0.2}, float64(0.3), false},
		{[]float64{0.1, 0.2}, float64(0.1), true},
		{[]int{}, 1, false},
		{[]int{1, 2}, int(3), false},
		{[]int{1, 2}, int(1), true},
		{[]int8{}, 1, false},
		{[]int8{1, 2}, int8(3), false},
		{[]int8{1, 2}, int8(1), true},
		{[]int16{}, 1, false},
		{[]int16{1, 2}, int16(3), false},
		{[]int16{1, 2}, int16(1), true},
		{[]int32{}, 1, false},
		{[]int32{1, 2}, int32(3), false},
		{[]int32{1, 2}, int32(1), true},
		{[]int64{}, 1, false},
		{[]int64{1, 2}, int64(3), false},
		{[]int64{1, 2}, int64(1), true},
		{[]uint{}, 1, false},
		{[]uint{1, 2}, uint(3), false},
		{[]uint{1, 2}, uint(1), true},
		{[]uint16{}, 1, false},
		{[]uint16{1, 2}, uint16(3), false},
		{[]uint16{1, 2}, uint16(1), true},
		{[]uint32{}, 1, false},
		{[]uint32{1, 2}, uint32(3), false},
		{[]uint32{1, 2}, uint32(1), true},
		{[]uint64{}, 1, false},
		{[]uint64{1, 2}, uint64(3), false},
		{[]uint64{1, 2}, uint64(1), true},
		{[]uintptr{}, 1, false},
		{[]uintptr{1, 2}, uintptr(3), false},
		{[]uintptr{1, 2}, uintptr(1), true},
		{[]string{}, "a", false},
		{[]string{"a", "b"}, "c", false},
		{[]string{"a", "b"}, "a", true},
	}

	for _, t := range tt {
		s.Equal(t.expected, s.support.ArrayContains(t.arr, t.val))
	}
}

func (s *SupportSuite) TestAESEncryptDecrypt() {
	ciphertext, err := s.support.AESEncrypt([]byte("!@#$%^&*()"), []byte("58f364f29b568807ab9cffa22c99b538"))
	s.NoError(err)

	_, err = s.support.AESDecrypt(ciphertext, []byte("58f364f29b568807ab9cffa22c99b583"))
	s.Error(err)

	key := []byte("58f364f29b568807ab9cffa22c99b538")
	ciphertext, err = s.support.AESEncrypt([]byte("!@#$%^&*()"), key)
	s.NoError(err)

	plaintext, err := s.support.AESDecrypt(ciphertext, key)
	s.NoError(err)
	s.Equal(plaintext, []byte("!@#$%^&*()"))

	_, err = s.support.AESEncrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")

	_, err = s.support.AESDecrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *SupportSuite) TestCaptureOutput() {
	output := s.support.CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	s.Equal("foobar", output)
}

func (s *SupportSuite) TestIsCamelCase() {
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
		s.Equal(t[1], s.support.IsCamelCase(t[0].(string)))
	}
}

func (s *SupportSuite) TestIsChainCase() {
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
		s.Equal(t[1], s.support.IsChainCase(t[0].(string)))
	}
}

func (s *SupportSuite) TestIsFlatCase() {
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
		s.Equal(t[1], s.support.IsFlatCase(t[0].(string)))
	}
}

func (s *SupportSuite) TestIsPascalCase() {
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
		s.Equal(t[1], s.support.IsPascalCase(t[0].(string)))
	}
}

func (s *SupportSuite) TestIsSnakeCase() {
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
		s.Equal(t[1], s.support.IsSnakeCase(t[0].(string)))
	}
}

func (s *SupportSuite) TestParseEnv() {
	type cfg1 struct {
		Admins  map[string]string `env:"TEST_ADMINS" envDefault:"user1:pass1,user2:pass2"`
		Hosts   []string          `env:"TEST_HOSTS" envDefault:"0.0.0.0,1.1.1.1"`
		Secret  []byte            `env:"TEST_SECRET" envDefault:"hello"`
		Secrets [][]byte          `env:"TEST_SECRETS" envDefault:"hello,world"`
	}

	c1 := &cfg1{}
	err := s.support.ParseEnv(c1)
	s.Nil(err)
	s.Equal(map[string]string{"user1": "pass1", "user2": "pass2"}, c1.Admins)
	s.Equal([]string{"0.0.0.0", "1.1.1.1"}, c1.Hosts)
	s.Equal([]byte("hello"), c1.Secret)
	s.Equal([][]byte{[]byte("hello"), []byte("world")}, c1.Secrets)

	type cfg2 struct {
		Users map[string]bool `env:"TEST_USERS" envDefault:"user1:true,user2:false"`
	}

	c2 := &cfg2{}
	err = s.support.ParseEnv(c2)
	s.NotNil(err)

	type cfg3 struct {
		Users map[string]string `env:"TEST_USERS" envDefault:"user1"`
	}

	c3 := &cfg3{}
	err = s.support.ParseEnv(c3)
	s.Nil(err)
	s.Equal(map[string]string{}, c3.Users)
}

func (s *SupportSuite) TestToCamelCase() {
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
		s.Equal(t[1], s.support.ToCamelCase(t[0]))
	}
}

func (s *SupportSuite) TestToChainCase() {
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
		s.Equal(t[1], s.support.ToChainCase(t[0]))
	}
}

func (s *SupportSuite) TestToFlatCase() {
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
		s.Equal(t[1], s.support.ToFlatCase(t[0]))
	}
}

func (s *SupportSuite) TestToPascalCase() {
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
		s.Equal(t[1], s.support.ToPascalCase(t[0]))
	}
}

func (s *SupportSuite) TestToSnakeCase() {
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
		s.Equal(t[1], s.support.ToSnakeCase(t[0]))
	}
}

func TestSupportSuite(t *testing.T) {
	appy.RunTestSuite(t, new(SupportSuite))
}
