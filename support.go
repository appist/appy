package appy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/env"
)

type (
	// Supporter satisfies Support type and implements all its functions, mainly used for unit testing's mock.
	Supporter interface {
		AESDecrypt(ciphertext []byte, key []byte) ([]byte, error)
		AESEncrypt(plaintext []byte, key []byte) ([]byte, error)
		ParseEnv(c interface{}) error
	}

	// Support contains the useful functions.
	Support struct{}
)

// AESDecrypt decrypts a cipher text into a plain text using the key with AES.
func (s *Support) AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

// AESEncrypt encrypts a plaintext into a cipher text using the key with AES.
func (s *Support) AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

type capturer struct {
	stdout bool
	stderr bool
}

func (c *capturer) capture(f func()) string {
	r, w, _ := os.Pipe()

	if c.stdout {
		stdout := os.Stdout
		os.Stdout = w
		defer func() {
			os.Stdout = stdout
		}()
	}

	if c.stderr {
		stderr := os.Stderr
		os.Stderr = w
		defer func() {
			os.Stderr = stderr
		}()
	}

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

// CaptureOutput captures stdout and stderr.
func CaptureOutput(f func()) string {
	capturer := &capturer{stdout: true, stderr: true}
	return capturer.capture(f)
}

// ParseEnv parses the environment variables into config struct.
func (s *Support) ParseEnv(c interface{}) error {
	err := env.ParseWithFuncs(c, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(map[string]string{}): func(v string) (interface{}, error) {
			newMaps := map[string]string{}
			maps := strings.Split(v, ",")
			for _, m := range maps {
				splits := strings.Split(m, ":")
				if len(splits) != 2 {
					continue
				}

				newMaps[splits[0]] = splits[1]
			}

			return newMaps, nil
		},
		reflect.TypeOf([]byte{}): func(v string) (interface{}, error) {
			return []byte(v), nil
		},
		reflect.TypeOf([][]byte{}): func(v string) (interface{}, error) {
			newBytes := [][]byte{}
			bytes := strings.Split(v, ",")
			for _, b := range bytes {
				newBytes = append(newBytes, []byte(b))
			}

			return newBytes, nil
		},
	})

	if err != nil {
		return err
	}

	return nil
}
