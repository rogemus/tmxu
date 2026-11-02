package tmux

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestBuildNewSessionCmd(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		dir         string
		want        []string
	}{
		{
			name:        "no dir leaves start directory to tmux",
			sessionName: "test-session",
			dir:         "",
			want:        []string{"new-session", "-d", "-s", "test-session"},
		},
		{
			name:        "dir adds -c",
			sessionName: "test-session",
			dir:         "/test/path",
			want:        []string{"new-session", "-d", "-s", "test-session", "-c", "/test/path"},
		},
		{
			name:        "name with spaces stays one argument",
			sessionName: "my session",
			dir:         "/test/my path",
			want:        []string{"new-session", "-d", "-s", "my session", "-c", "/test/my path"},
		},
		{
			name:        "values are not quoted, argv is passed without a shell",
			sessionName: "it's",
			dir:         "",
			want:        []string{"new-session", "-d", "-s", "it's"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewSessionCmd(tt.sessionName, tt.dir)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildNewSessionCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewSessionError(t *testing.T) {
	cause := &exec.ExitError{}

	t.Run("duplicate session maps to ErrSessionExists", func(t *testing.T) {
		got := newSessionError("a", []byte("duplicate session: a\n"), cause)

		if !errors.Is(got, ErrSessionExists) {
			t.Errorf("errors.Is(%v, ErrSessionExists) = false", got)
		}
		if errors.Is(got, ErrCannotCreateSession) {
			t.Errorf("a duplicate must not report ErrCannotCreateSession: %v", got)
		}
		if !strings.Contains(got.Error(), "a") {
			t.Errorf("error %q does not name the session", got)
		}
	})

	t.Run("other failure keeps tmux message and the cause", func(t *testing.T) {
		got := newSessionError("a", []byte("no server running\n"), cause)

		if !errors.Is(got, ErrCannotCreateSession) {
			t.Errorf("errors.Is(%v, ErrCannotCreateSession) = false", got)
		}
		if errors.Is(got, ErrSessionExists) {
			t.Errorf("a generic failure must not report ErrSessionExists: %v", got)
		}
		if !errors.Is(got, cause) {
			t.Errorf("errors.Is(%v, cause) = false", got)
		}
		if !strings.Contains(got.Error(), "no server running") {
			t.Errorf("error %q lost tmux's message", got)
		}
	})

	t.Run("empty output still gives a usable error", func(t *testing.T) {
		got := newSessionError("a", nil, cause)

		if !errors.Is(got, ErrCannotCreateSession) {
			t.Errorf("errors.Is(%v, ErrCannotCreateSession) = false", got)
		}
	})
}
