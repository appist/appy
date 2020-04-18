package support

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// AESDecrypt decrypts a cipher text into a plain text using the key with
// AES-256 algorithm.
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	decodedKey, _ := hex.DecodeString(string(key))
	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// AESEncrypt encrypts a plaintext into a cipher text using the key with
// AES-256 algorithm.
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	decodedKey, _ := hex.DecodeString(string(key))
	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// GenerateRandomBytes generates random bytes of the specific length.
func GenerateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)

	return bytes
}
