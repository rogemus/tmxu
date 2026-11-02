package utils

import "strings"

func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func Escape(s string) string {
	return strings.ReplaceAll(s, "#", "##")
}
