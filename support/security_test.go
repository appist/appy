package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type SecuritySuite struct {
	test.Suite
}

func (s *SecuritySuite) SetupTest() {
}

func (s *SecuritySuite) TearDownTest() {
}

func (s *SecuritySuite) TestEncryptInvalidKeyLength() {
	_, err := Encrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *SecuritySuite) TestDecryptInvalidKeyLength() {
	_, err := Decrypt([]byte("dummy"), []byte("key"))
	s.EqualError(err, "crypto/aes: invalid key size 3")
}

func (s *SecuritySuite) TestEncryptDecryptWithValidKey() {
	var err error
	key := []byte("58f364f29b568807ab9cffa22c99b538")
	ciphertext, err := Encrypt([]byte("!@#$%^&*()"), key)
	s.NoError(err)

	plaintext, err := Decrypt(ciphertext, key)
	s.NoError(err)
	s.Equal(plaintext, []byte("!@#$%^&*()"))
}

func (s *SecuritySuite) TestEncryptDecryptWithInvalidKey() {
	var err error
	ciphertext, err := Encrypt([]byte("!@#$%^&*()"), []byte("58f364f29b568807ab9cffa22c99b538"))
	s.NoError(err)

	_, err = Decrypt(ciphertext, []byte("58f364f29b568807ab9cffa22c99b583"))
	s.Error(err)
}

func TestSecurity(t *testing.T) {
	test.Run(t, new(SecuritySuite))
}
