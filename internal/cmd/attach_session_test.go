package cmd

import (
	"errors"
	"testing"
)

func TestAttachSessionRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{name: "no arguments", args: nil, want: ErrSessionNameNotProvided},
		{name: "menu flag without a name", args: []string{"-menu"}, want: ErrSessionNameNotProvided},
		{name: "unknown flag", args: []string{"-bad", "a"}, want: ErrCannotParseAttachCmdFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, out, _ := testIOs("")

			err := NewAttachSession(testConfig(), ios).Run(tt.args)

			if !errors.Is(err, tt.want) {
				t.Fatalf("Run(%q) error = %v, want %v", tt.args, err, tt.want)
			}
			if !errors.Is(err, ErrUsage) {
				t.Errorf("error %v does not wrap ErrUsage, main would not exit 2", err)
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing on a usage error", out)
			}
		})
	}
}
