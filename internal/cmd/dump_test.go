package cmd

import (
	"errors"
	"testing"
)

// Only invocations that fail before tmux is called are tested here.
func TestDumpRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"-bad"}},
		{name: "flag with a value it does not take", args: []string{"-json=maybe"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, out, _ := testIOs("")

			err := NewDump(testConfig(), ios).Run(tt.args)

			if !errors.Is(err, ErrCannotParseDumpCmdFlags) {
				t.Fatalf("Run(%q) error = %v, want %v", tt.args, err, ErrCannotParseDumpCmdFlags)
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
