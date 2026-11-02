package cmd

import (
	"errors"
	"testing"

	"github.com/rogemus/tmxu/internal/config"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestVersionRun(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "plain", version: "1.2.3", want: "Current tmxu version 1.2.3\n"},
		{name: "empty version", version: "", want: "Current tmxu version \n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, out, errOut := testIOs("")

			err := NewVersion(config.Config{Version: tt.version}, ios).Run(nil)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if out.String() != tt.want {
				t.Errorf("stdout = %q, want %q", out, tt.want)
			}
			if errOut.Len() != 0 {
				t.Errorf("stderr = %q, want nothing", errOut)
			}
		})
	}
}

func TestVersionRunReturnsWriteError(t *testing.T) {
	boom := errors.New("closed pipe")
	ios, _, _ := testIOs("")
	ios.Out = failingWriter{boom}

	err := NewVersion(testConfig(), ios).Run(nil)

	if !errors.Is(err, boom) {
		t.Errorf("Run() error = %v, want %v", err, boom)
	}
}
