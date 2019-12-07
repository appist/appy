package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type CryptoSuite struct {
	test.Suite
}

func (s *CryptoSuite) SetupTest() {
}

func (s *CryptoSuite) TearDownTest() {
}

func (s *CryptoSuite) TestAESEncryptInvalidKeyLength() {
	_, err := AESEncrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *CryptoSuite) TestAESDecryptInvalidKeyLength() {
	_, err := AESDecrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *CryptoSuite) TestAESEncryptAESDecryptWithValidKey() {
	var err error
	key := []byte("58f364f29b568807ab9cffa22c99b538")
	ciphertext, err := AESEncrypt([]byte("!@#$%^&*()"), key)
	s.NoError(err)

	plaintext, err := AESDecrypt(ciphertext, key)
	s.NoError(err)
	s.Equal(plaintext, []byte("!@#$%^&*()"))
}

func (s *CryptoSuite) TestAESEncryptAESDecryptWithInvalidKey() {
	var err error
	ciphertext, err := AESEncrypt([]byte("!@#$%^&*()"), []byte("58f364f29b568807ab9cffa22c99b538"))
	s.NoError(err)

	_, err = AESDecrypt(ciphertext, []byte("58f364f29b568807ab9cffa22c99b583"))
	s.Error(err)
}

func TestCryptoSuite(t *testing.T) {
	test.RunSuite(t, new(CryptoSuite))
}
