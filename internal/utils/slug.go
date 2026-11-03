package utils

import (
	"strings"
	"unicode"
)

func Slug(in string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}

		return '_'
	}, in)
}
