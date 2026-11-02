package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestNewSessionRunUsageErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        error
		wantMessage string
	}{
		{name: "no arguments", args: nil, want: ErrSessionNameNotProvided},
		{name: "empty name", args: []string{""}, want: ErrSessionNameNotProvided},
		{name: "only flags", args: []string{"-path", "/tmp"}, want: ErrSessionNameNotProvided},
		{
			name:        "two names",
			args:        []string{"a", "b"},
			want:        ErrUnexpectedArgument,
			wantMessage: `"b"`,
		},
		{
			name:        "flag after the name is reported",
			args:        []string{"a", "-path", "/tmp"},
			want:        ErrUnexpectedArgument,
			wantMessage: `"-path"`,
		},
		{name: "unknown flag", args: []string{"-bad", "a"}, want: ErrCannotParseNewSessionCmdFlags},
		{name: "flag without value", args: []string{"-path"}, want: ErrCannotParseNewSessionCmdFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, out, _ := testIOs("")

			err := NewNewSession(testConfig(), ios).Run(tt.args)

			if !errors.Is(err, tt.want) {
				t.Fatalf("Run(%q) error = %v, want %v", tt.args, err, tt.want)
			}
			if !errors.Is(err, ErrUsage) {
				t.Errorf("error %v does not wrap ErrUsage, main would not exit 2", err)
			}
			if tt.wantMessage != "" && !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("error %q does not contain %s", err, tt.wantMessage)
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing on a usage error", out)
			}
		})
	}
}
