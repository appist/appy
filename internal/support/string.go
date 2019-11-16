package support

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

// IsPascalCase checks if a string is a PascalCase.
func IsPascalCase(str string) bool {
	if isFirstRuneDigit(str) {
		return false
	}

	return isAlphanumeric(str) && isFirstRuneUpper(str)
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
