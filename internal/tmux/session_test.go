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
			name: "single session",
			sessions: []Session{
				{Name: "test-session", FirstWindowName: "test-window", FirstWindowPath: "/test/path"},
			},
			want: "new-session -d -s test-session -n test-window -c /test/path",
		},
		{
			name: "two sessions are joined with a separator",
			sessions: []Session{
				{Name: "api", FirstWindowName: "editor", FirstWindowPath: "/code/api"},
				{Name: "web", FirstWindowName: "shell", FirstWindowPath: "/code/web"},
			},
			want: "new-session -d -s api -n editor -c /code/api ; new-session -d -s web -n shell -c /code/web",
		},
		{
			name: "three sessions keep order",
			sessions: []Session{
				{Name: "a", FirstWindowName: "w", FirstWindowPath: "/a"},
				{Name: "b", FirstWindowName: "w", FirstWindowPath: "/b"},
				{Name: "c", FirstWindowName: "w", FirstWindowPath: "/c"},
			},
			want: "new-session -d -s a -n w -c /a ; new-session -d -s b -n w -c /b ; new-session -d -s c -n w -c /c",
		},
		{
			name: "ids are not part of the command",
			sessions: []Session{
				{ID: "test-session-id", Name: "api", FirstWindowName: "editor", FirstWindowPath: "/code/api"},
			},
			want: "new-session -d -s api -n editor -c /code/api",
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
