package support

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

// IsCamelCase checks if a string is camelCase.
func IsCamelCase(s string) bool {
	if isFirstRuneDigit(s) {
		return false
	}

	return isMadeByAlphanumeric(s) && isFirstRuneLower(s)
}

// IsChainCase checks if a string is a chain-case.
func IsChainCase(s string) bool {
	if strings.Contains(s, "-") {
		fields := strings.Split(s, "-")
		for _, field := range fields {
			if !isMadeByLowerAndDigit(field) {
				return false
			}
		}
		return true
	}

	return isMadeByLowerAndDigit(s)
}

// IsFlatCase checks if a string is a flatcase.
func IsFlatCase(s string) bool {
	if isFirstRuneDigit(s) {
		return false
	}

	return isMadeByLowerAndDigit(s)
}

// IsPascalCase checks if a string is a PascalCase.
func IsPascalCase(str string) bool {
	if isFirstRuneDigit(str) {
		return false
	}

	return isAlphanumeric(str) && isFirstRuneUpper(str)
}

// IsSnakeCase checks if a string is a snake_case.
func IsSnakeCase(s string) bool {
	if strings.Contains(s, "_") {
		fields := strings.Split(s, "_")
		for _, field := range fields {
			if !isMadeByLowerAndDigit(field) {
				return false
			}
		}

		return true
	}

	return isMadeByLowerAndDigit(s)
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

// ToChainCase converts a string to chain-case style.
func ToChainCase(s string) string {
	if len(s) == 0 {
		return s
	}

	fields := splitToLowerFields(s)
	return strings.Join(fields, "-")
}

// ToFlatCase converts a string to flatcase style.
func ToFlatCase(s string) string {
	if len(s) == 0 {
		return s
	}

	fields := splitToLowerFields(s)
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

func getRuneAt(s string, i int) rune {
	if len(s) == 0 {
		return 0
	}

	rs := []rune(s)
	return rs[0]
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

func isFirstRuneLower(s string) bool {
	if len(s) == 0 {
		return false
	}

	return unicode.IsLower(getRuneAt(s, 0))
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
