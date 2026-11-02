package cmd

import (
	"strings"
	"testing"
)

func TestSaveSessionsRunDeclined(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "n", input: "n\n"},
		{name: "empty line takes the default", input: "\n"},
		{name: "unrelated text", input: "later\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, out, errOut := testIOs(tt.input)

			err := NewSaveSessions(testConfig(), ios).Run(nil)

			if err != nil {
				t.Fatalf("Run() error = %v, want nil when the user declines", err)
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing", out)
			}
			if !strings.Contains(errOut.String(), "[y/N]") {
				t.Errorf("stderr = %q, want the prompt", errOut)
			}
			if !strings.HasSuffix(errOut.String(), "Aborted\n") {
				t.Errorf("stderr = %q, want it to end with %q", errOut, "Aborted\n")
			}
		})
	}
}
