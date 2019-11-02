package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type SupportSuite struct {
	appy.TestSuite
}

func (s *SupportSuite) SetupTest() {
}

func (s *SupportSuite) TearDownTest() {
}

func (s *SupportSuite) TestAESEncryptInvalidKeyLength() {
	_, err := appy.AESEncrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *SupportSuite) TestAESDecryptInvalidKeyLength() {
	_, err := appy.AESDecrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *SupportSuite) TestAESEncryptAESDecryptWithValidKey() {
	var err error
	key := []byte("58f364f29b568807ab9cffa22c99b538")
	ciphertext, err := appy.AESEncrypt([]byte("!@#$%^&*()"), key)
	s.NoError(err)

	plaintext, err := appy.AESDecrypt(ciphertext, key)
	s.NoError(err)
	s.Equal(plaintext, []byte("!@#$%^&*()"))
}

func (s *SupportSuite) TestAESEncryptAESDecryptWithInvalidKey() {
	var err error
	ciphertext, err := appy.AESEncrypt([]byte("!@#$%^&*()"), []byte("58f364f29b568807ab9cffa22c99b538"))
	s.NoError(err)

	_, err = appy.AESDecrypt(ciphertext, []byte("58f364f29b568807ab9cffa22c99b583"))
	s.Error(err)
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
		s.Equal(t.expected, appy.ArrayContains(t.arr, t.val))
	}
}

func (s *SupportSuite) TestDeepClone() {
	type User struct {
		Email string
		Name  string
	}

	type Employee struct {
		Email string
		Name  string
		Role  string
	}

	user := User{Email: "john_doe@gmail.com", Name: "John Doe"}
	employee := Employee{}
	appy.DeepClone(&employee, &user)
	s.Equal("john_doe@gmail.com", employee.Email)
	s.Equal("John Doe", employee.Name)

	employees := []Employee{}
	appy.DeepClone(&employees, &user)
	s.Equal(1, len(employees))
	s.Equal("john_doe@gmail.com", employees[0].Email)
	s.Equal("John Doe", employees[0].Name)

	users := []User{
		{Email: "john_doe1@gmail.com", Name: "John Doe 1"},
		{Email: "john_doe2@gmail.com", Name: "John Doe 2"},
	}
	employees = []Employee{}
	appy.DeepClone(&employees, &users)
	s.Equal(2, len(employees))
	s.Equal("john_doe1@gmail.com", employees[0].Email)
	s.Equal("John Doe 1", employees[0].Name)
	s.Equal("john_doe2@gmail.com", employees[1].Email)
	s.Equal("John Doe 2", employees[1].Name)
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
		s.Equal(t[1], appy.IsPascalCase(t[0].(string)))
	}
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
		s.Equal(t[1], appy.ToCamelCase(t[0]))
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
		s.Equal(t[1], appy.ToSnakeCase(t[0]))
	}
}

func (s *SupportSuite) TestParseEnvWithSupportedTypes() {
	type testConfig struct {
		Admins  map[string]string `env:"TEST_ADMINS" envDefault:"user1:pass1,user2:pass2"`
		Hosts   []string          `env:"TEST_HOSTS" envDefault:"0.0.0.0,1.1.1.1"`
		Secret  []byte            `env:"TEST_SECRET" envDefault:"hello"`
		Secrets [][]byte          `env:"TEST_SECRETS" envDefault:"hello,world"`
	}

	c := &testConfig{}
	appy.ParseEnv(c)
	s.Equal(map[string]string{"user1": "pass1", "user2": "pass2"}, c.Admins)
	s.Equal([]string{"0.0.0.0", "1.1.1.1"}, c.Hosts)
	s.Equal([]byte("hello"), c.Secret)
	s.Equal([][]byte{[]byte("hello"), []byte("world")}, c.Secrets)
}

func (s *SupportSuite) TestParseEnvWithUnsupportedTypes() {
	type testConfig struct {
		Users map[string]int `env:"TEST_USERS" envDefault:"user1:1,user2:2"`
	}

	err := appy.ParseEnv(&testConfig{})
	s.NotNil(err)
}

func (s *SupportSuite) TestParseEnvWithInvalidFormat() {
	type testConfig struct {
		Users map[string]string `env:"TEST_USERS" envDefault:"user1"`
	}

	c := &testConfig{}
	appy.ParseEnv(c)
	s.Equal(map[string]string{}, c.Users)
}

func TestSupportSuite(t *testing.T) {
	appy.RunTestSuite(t, new(SupportSuite))
}
