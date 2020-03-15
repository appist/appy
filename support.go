package appy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/caarlos0/env"
	"github.com/fatih/camelcase"
)

type (
	// Supporter satisfies Support type and implements all its functions, mainly used for mocking in unit test.
	Supporter interface {
		ArrayContains(arr interface{}, val interface{}) bool
		AESDecrypt(ciphertext []byte, key []byte) ([]byte, error)
		AESEncrypt(plaintext []byte, key []byte) ([]byte, error)
		CaptureOutput(f func()) string
		IsCamelCase(s string) bool
		IsChainCase(s string) bool
		IsFlatCase(s string) bool
		IsPascalCase(s string) bool
		IsSnakeCase(s string) bool
		ParseEnv(c interface{}) error
		ToCamelCase(str string) string
		ToChainCase(str string) string
		ToFlatCase(str string) string
		ToPascalCase(str string) string
		ToSnakeCase(str string) string
	}

	// Support contains the useful functions.
	Support struct{}
)

// ArrayContains checks if a value is in a slice of the same type.
func (s *Support) ArrayContains(arr interface{}, val interface{}) bool {
	arrT := reflect.TypeOf(arr)
	valT := reflect.TypeOf(val)
	if (arrT.Kind().String() != "array" && arrT.Kind().String() != "slice") ||
		arrT.Elem().String() != valT.Kind().String() {
		return false
	}

	switch arr := arr.(type) {
	case []bool:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []byte:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []complex64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []complex128:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []float32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []float64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int8:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int16:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint16:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint32:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uint64:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []uintptr:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []string:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	}

	return false
}

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
func (s *Support) CaptureOutput(f func()) string {
	capturer := &capturer{stdout: true, stderr: true}
	return capturer.capture(f)
}

// IsCamelCase checks if a string is camelCase.
func (s *Support) IsCamelCase(str string) bool {
	return !isFirstRuneDigit(str) && isMadeByAlphanumeric(str) && unicode.IsLower(runeAt(str, 0))
}

// IsChainCase checks if a string is a chain-case.
func (s *Support) IsChainCase(str string) bool {
	if strings.Contains(str, "-") {
		fields := strings.Split(str, "-")
		for _, field := range fields {
			if !isMadeByLowerAndDigit(field) {
				return false
			}
		}

		return true
	}

	return isMadeByLowerAndDigit(str)
}

// IsFlatCase checks if a string is a flatcase.
func (s *Support) IsFlatCase(str string) bool {
	return !isFirstRuneDigit(str) && isMadeByLowerAndDigit(str)
}

// IsPascalCase checks if a string is a PascalCase.
func (s *Support) IsPascalCase(str string) bool {
	if isFirstRuneDigit(str) {
		return false
	}

	return isAlphanumeric(str) && unicode.IsUpper(runeAt(str, 0))
}

// IsSnakeCase checks if a string is a snake_case.
func (s *Support) IsSnakeCase(str string) bool {
	if strings.Contains(str, "_") {
		fields := strings.Split(str, "_")
		for _, field := range fields {
			if !isMadeByLowerAndDigit(field) {
				return false
			}
		}

		return true
	}

	return isMadeByLowerAndDigit(str)
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
		reflect.TypeOf(map[string]int{}): func(v string) (interface{}, error) {
			newMaps := map[string]int{}
			maps := strings.Split(v, ",")
			for _, m := range maps {
				splits := strings.Split(m, ":")
				if len(splits) != 2 {
					continue
				}

				val, _ := strconv.Atoi(splits[1])
				newMaps[splits[0]] = val
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

// ToCamelCase converts a string to camelCase style.
func (s *Support) ToCamelCase(str string) string {
	if len(str) == 0 {
		return str
	}

	fields := splitToLowerFields(str)
	for i, f := range fields {
		if i != 0 {
			fields[i] = toUpperFirstRune(f)
		}
	}
	return strings.Join(fields, "")
}

// ToChainCase converts a string to chain-case style.
func (s *Support) ToChainCase(str string) string {
	if len(str) == 0 {
		return str
	}

	fields := splitToLowerFields(str)
	return strings.Join(fields, "-")
}

// ToFlatCase converts a string to flatcase style.
func (s *Support) ToFlatCase(str string) string {
	if len(str) == 0 {
		return str
	}

	fields := splitToLowerFields(str)
	return strings.Join(fields, "")
}

// ToPascalCase converts a string to PascalCase style.
func (s *Support) ToPascalCase(str string) string {
	if len(str) == 0 {
		return str
	}

	fields := splitToLowerFields(str)
	for i, f := range fields {
		fields[i] = toUpperFirstRune(f)
	}

	return strings.Join(fields, "")
}

// ToSnakeCase converts a string to snake_case style.
func (s *Support) ToSnakeCase(str string) string {
	if len(str) == 0 {
		return str
	}

	fields := splitToLowerFields(str)
	return strings.Join(fields, "_")
}

func isAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !unicode.IsUpper(r) && !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isFirstRuneDigit(s string) bool {
	if len(s) == 0 {
		return false
	}

	return unicode.IsDigit(runeAt(s, 0))
}

func isMadeByAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !unicode.IsUpper(r) && !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isMadeByLowerAndDigit(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func runeAt(s string, i int) rune {
	rs := []rune(s)
	return rs[0]
}

func splitToLowerFields(s string) []string {
	defaultCap := len([]rune(s)) / 3
	fields := make([]string, 0, defaultCap)

	for _, sf := range strings.Fields(s) {
		for _, su := range strings.Split(sf, "_") {
			for _, sh := range strings.Split(su, "-") {
				for _, sc := range camelcase.Split(sh) {
					fields = append(fields, strings.ToLower(sc))
				}
			}
		}
	}
	return fields
}

func toUpperFirstRune(s string) string {
	rs := []rune(s)
	return strings.ToUpper(string(rs[0])) + string(rs[1:])
}
