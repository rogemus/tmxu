package tmux

import "testing"

func TestBuildNewSessionsCmd(t *testing.T) {
	tests := []struct {
		name     string
		sessions []Session
		want     string
	}{
		{
			name:     "no sessions",
			sessions: nil,
			want:     "",
		},
		{
			name:     "single session",
			sessions: []Session{{Name: "test-session"}},
			want:     "new-session -d -s 'test-session'",
		},
		{
			name:     "two sessions are joined with a separator",
			sessions: []Session{{Name: "api"}, {Name: "web"}},
			want:     "new-session -d -s 'api' ; new-session -d -s 'web'",
		},
		{
			name:     "three sessions keep order",
			sessions: []Session{{Name: "a"}, {Name: "b"}, {Name: "c"}},
			want:     "new-session -d -s 'a' ; new-session -d -s 'b' ; new-session -d -s 'c'",
		},
		{
			name:     "id is not part of the command",
			sessions: []Session{{ID: "test-session-id", Name: "api"}},
			want:     "new-session -d -s 'api'",
		},
		{
			name:     "name with a space",
			sessions: []Session{{Name: "test session"}},
			want:     "new-session -d -s 'test session'",
		},
		{
			name:     "name with a single quote",
			sessions: []Session{{Name: "it's"}},
			want:     `new-session -d -s 'it'\''s'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewSessionsCmd(tt.sessions)
			if got != tt.want {
				t.Errorf("buildNewSessionsCmd() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestSessionStrName(t *testing.T) {
	got := Session{Name: "test-session"}.GetName()
	if want := "'test-session'"; got != want {
		t.Errorf("StrName() = %s, want %s", got, want)
	}
}

func TestSessionEq(t *testing.T) {
	s := Session{ID: "test-session-id"}

	if !s.Eq(Session{ID: "test-session-id"}) {
		t.Error("sessions with the same ID should be equal")
	}

	if s.Eq(Session{ID: "other-session-id"}) {
		t.Error("sessions with different IDs should not be equal")
	}

	if s.Eq(Window{ID: "test-session-id"}) {
		t.Error("a Window should never equal a Session")
	}
}
