package appy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"reflect"
	"strings"
	"unicode"

	"github.com/caarlos0/env"
	"github.com/fatih/camelcase"
	"github.com/jinzhu/copier"
)

// AESDecrypt decrypts a cipher text into a plain text using the key with AES.
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

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
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// ArrayContains checks if a value is in a slice of the same type.
func ArrayContains(arr interface{}, val interface{}) bool {
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

// DeepClone deeply clones from 1 interface to another.
func DeepClone(dst, src interface{}) error {
	return copier.Copy(dst, src)
}

// IsPascalCase checks if a string is a PascalCase.
func IsPascalCase(str string) bool {
	if isFirstRuneDigit(str) {
		return false
	}

	return isAlphanumeric(str) && isFirstRuneUpper(str)
}

// ParseEnv parses the environment variables into the config.
func ParseEnv(c interface{}) error {
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

// ToCamelCase converts a string to camelCase style.
func ToCamelCase(str string) string {
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

// ToSnakeCase converts a string to snake_case style.
func ToSnakeCase(str string) string {
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

func isFirstRuneUpper(s string) bool {
	if len(s) == 0 {
		return false
	}

	return unicode.IsUpper(runeAt(s, 0))
}

func runeAt(s string, i int) rune {
	if len(s) == 0 {
		return 0
	}

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
