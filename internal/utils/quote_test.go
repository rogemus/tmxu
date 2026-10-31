package utils

import "testing"

func TestQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain text", "test-value", `'test-value'`},
		{"space", "test value", `'test value'`},
		{"empty string", "", `''`},
		{"single quote", "it's", `'it'\''s'`},
		{"two single quotes", "a'b'c", `'a'\''b'\''c'`},
		{"double quotes and variable stay literal", `say "hi" $HOME`, `'say "hi" $HOME'`},
		{"backslash stays literal", `a\b`, `'a\b'`},
		{"semicolon stays literal", "a;b", `'a;b'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Quote(tt.in)
			if got != tt.want {
				t.Errorf("Quote(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}
