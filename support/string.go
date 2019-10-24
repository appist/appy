package support

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

// IsPascalCase checks if a string is a PascalCase.
func IsPascalCase(s string) bool {
	if isFirstRuneDigit(s) {
		return false
	}

	return isAlphanumeric(s) && isFirstRuneUpper(s)
}

// ToSnakeCase converts a string to snake_case style.
func ToSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	fields := splitToLowerFields(s)
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
