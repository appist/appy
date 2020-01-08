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
	output := appy.CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	s.Equal("foobar", output)
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
		Users map[string]int `env:"TEST_USERS" envDefault:"user1:1,user2:2"`
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

func TestSupportSuite(t *testing.T) {
	appy.RunTestSuite(t, new(SupportSuite))
}
