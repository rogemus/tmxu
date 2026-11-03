package utils

import (
	"testing"

	"github.com/rogemus/tmxu/internal/expect"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{name: "empty", in: ``, expected: ``},
		{name: "plain", in: `work`, expected: `work`},
		{name: "uppercase kept", in: `Work`, expected: `Work`},
		{name: "space", in: `my work`, expected: `my_work`},
		{name: "repeated spaces", in: `a   b`, expected: `a___b`},
		{name: "leading trailing spaces", in: `  work  `, expected: `__work__`},
		{name: "format", in: `#S`, expected: `_S`},
		{name: "command", in: `#(date)`, expected: `__date_`},
		{name: "braces", in: `#{pane_id}`, expected: `__pane_id_`},
		{name: "dot and colon", in: `a.b:c`, expected: `a_b_c`},
		{name: "path", in: `~/dev/a b`, expected: `__dev_a_b`},
		{name: "hyphen kept", in: `my-work`, expected: `my-work`},
		{name: "underscore kept", in: `my_work`, expected: `my_work`},
		{name: "digits", in: `v2 build 10`, expected: `v2_build_10`},
		{name: "apostrophe", in: `it's`, expected: `it_s`},
		{name: "only specials", in: `!!!`, expected: `___`},
		{name: "accents kept", in: `Zażółć`, expected: `Zażółć`},
		{name: "non-latin kept", in: `日本語`, expected: `日本語`},
		{name: "emoji", in: `work 🚀`, expected: `work__`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Slug(test.in)
			expect.Eq(t, out, test.expected)
		})
	}
}
