package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type cryptoSuite struct {
	test.Suite
}

func (s *cryptoSuite) TestAESDecrypt() {
	{
		decrypted, err := AESDecrypt([]byte("foobar"), []byte("1234"))

		s.Equal("crypto/aes: invalid key size 4", err.Error())
		s.Nil(decrypted)
	}

	{
		decrypted, err := AESDecrypt([]byte("foobar"), []byte("58f364f29b568807ab9cffa22c99b538"))

		s.Nil(err)
		s.Nil(decrypted)
	}

	{
		decrypted, err := AESDecrypt([]byte("6e112491616f13ac0155ad17689d6fc685c69f150317c9eadc85a9ade35aff6154e387"), []byte("58f364f29b568807ab9cffa22c99b538"))

		s.Equal("cipher: message authentication failed", err.Error())
		s.Nil(decrypted)
	}
}

func (s *cryptoSuite) TestAESEncrypt() {
	{
		encrypted, err := AESEncrypt([]byte("foobar"), []byte("1234"))

		s.Equal("crypto/aes: invalid key size 4", err.Error())
		s.Nil(encrypted)
	}

	{
		encrypted, err := AESEncrypt([]byte("0.0.0.0"), []byte("58f364f29b568807ab9cffa22c99b538"))

		s.Nil(err)
		s.NotEmpty(encrypted)
	}
}

func (s *cryptoSuite) TestAESOps() {
	target := []byte("0.0.0.0")
	key := []byte("58f364f29b568807ab9cffa22c99b538")

	encrypted, err := AESEncrypt(target, key)

	s.Nil(err)
	s.NotEmpty(encrypted)

	decrypted, err := AESDecrypt(encrypted, key)

	s.Nil(err)
	s.Equal(target, decrypted)
}

func TestCryptoSuite(t *testing.T) {
	test.Run(t, new(cryptoSuite))
}
