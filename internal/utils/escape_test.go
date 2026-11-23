package utils

import (
	"testing"

	"github.com/rogemus/tmxu/internal/expect"
)

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
