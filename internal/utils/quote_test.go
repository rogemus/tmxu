package utils

import (
	"testing"

	"github.com/rogemus/tmxu/internal/expect"
)

func TestQuote(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{name: "empty", in: ``, expected: `''`},
		{name: "plain", in: `work`, expected: `'work'`},
		{name: "space", in: `my work`, expected: `'my work'`},
		{name: "semicolon", in: `a;b`, expected: `'a;b'`},
		{name: "dollar", in: `$HOME`, expected: `'$HOME'`},
		{name: "hash", in: `#S`, expected: `'#S'`},
		{name: "backslash", in: `a\b`, expected: `'a\b'`},
		{name: "unicode", in: `zażółć`, expected: `'zażółć'`},
		{name: "newline", in: "a\nb", expected: "'a\nb'"},
		{name: "single quote", in: `it's`, expected: `'it'\''s'`},
		{name: "only a quote", in: `'`, expected: `''\'''`},
		{name: "two quotes", in: `''`, expected: `''\'''\'''`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Quote(test.in)
			expect.Eq(t, out, test.expected)
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{name: "empty", in: ``, expected: ``},
		{name: "plain", in: `work`, expected: `work`},
		{name: "format", in: `#S`, expected: `##S`},
		{name: "command", in: `#(date)`, expected: `##(date)`},
		{name: "braces", in: `#{pane_id}`, expected: `##{pane_id}`},
		{name: "only hash", in: `#`, expected: `##`},
		{name: "hash in middle", in: `a#b`, expected: `a##b`},
		{name: "already escaped", in: `##`, expected: `####`},
		{name: "quote untouched", in: `it's`, expected: `it's`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Escape(test.in)
			expect.Eq(t, out, test.expected)
		})
	}
}

func TestQuoteEscape(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{name: "empty", in: ``, expected: `''`},
		{name: "plain", in: `work`, expected: `'work'`},
		{name: "space", in: `my work`, expected: `'my work'`},
		{name: "format", in: `#S`, expected: `'##S'`},
		{name: "command", in: `#(date)`, expected: `'##(date)'`},
		{name: "single quote", in: `it's`, expected: `'it'\''s'`},
		{name: "quote and hash", in: `it's #S`, expected: `'it'\''s ##S'`},
		{name: "dollar and hash", in: `$HOME #S`, expected: `'$HOME ##S'`},
		{name: "hash between quotes", in: `'#'`, expected: `''\''##'\'''`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Quote(Escape(test.in))
			expect.Eq(t, out, test.expected)
		})
	}
}
