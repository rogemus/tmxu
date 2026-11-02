package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "y", input: "y\n", want: true},
		{name: "yes", input: "yes\n", want: true},
		{name: "uppercase Y", input: "Y\n", want: true},
		{name: "uppercase YES", input: "YES\n", want: true},
		{name: "surrounding spaces", input: "  yes  \n", want: true},
		{name: "answer without newline", input: "y", want: true},
		{name: "n", input: "n\n", want: false},
		{name: "no", input: "no\n", want: false},
		{name: "empty line takes the default", input: "\n", want: false},
		{name: "unrelated text", input: "maybe\n", want: false},
		{name: "only the first line counts", input: "n\ny\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			got := Confirm("Continue?", strings.NewReader(tt.input), &out)
			if got != tt.want {
				t.Errorf("Confirm(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfirmWritesPrompt(t *testing.T) {
	var out bytes.Buffer

	Confirm("Save all?", strings.NewReader("n\n"), &out)

	if want := "Save all? [y/N]: "; out.String() != want {
		t.Errorf("prompt = %q, want %q", out.String(), want)
	}
}
