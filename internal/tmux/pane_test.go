package tmux

import (
	"testing"

	"github.com/rogemus/tmxu/internal/utils"
)

func TestPaneQuoting(t *testing.T) {
	tests := []struct {
		name      string
		pane      Pane
		wantTitle string
		wantPath  string
	}{
		{
			name:      "plain values",
			pane:      Pane{Title: "test-title", Path: "/test/path"},
			wantTitle: "'test-title'",
			wantPath:  "'/test/path'",
		},
		{
			name:      "spaces",
			pane:      Pane{Title: "test title", Path: "/test/my path"},
			wantTitle: "'test title'",
			wantPath:  "'/test/my path'",
		},
		{
			name:      "single quote is escaped",
			pane:      Pane{Title: "it's", Path: "/it's"},
			wantTitle: `'it'\''s'`,
			wantPath:  `'/it'\''s'`,
		},
		{
			name:      "dollar sign stays inside quotes",
			pane:      Pane{Title: "$HOME", Path: "/$HOME"},
			wantTitle: "'$HOME'",
			wantPath:  "'/$HOME'",
		},
		{
			name:      "empty values",
			pane:      Pane{},
			wantTitle: "''",
			wantPath:  "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pane.GetTitle(); got != tt.wantTitle {
				t.Errorf("GetTitle() = %s, want %s", got, tt.wantTitle)
			}
			if got := tt.pane.GetPath(); got != tt.wantPath {
				t.Errorf("GetPath() = %s, want %s", got, tt.wantPath)
			}
		})
	}
}

func TestPaneEq(t *testing.T) {
	p := Pane{ID: "%1", Title: "a"}

	tests := []struct {
		name  string
		other utils.Comperable
		want  bool
	}{
		{name: "same id", other: Pane{ID: "%1", Title: "other title"}, want: true},
		{name: "different id", other: Pane{ID: "%2", Title: "a"}, want: false},
		{name: "different type with same id", other: Window{ID: "%1"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.Eq(tt.other); got != tt.want {
				t.Errorf("Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildSplitPaneCmd(t *testing.T) {
	window := Window{Index: 2, SessionName: "test-session"}

	tests := []struct {
		name string
		pane Pane
		want string
	}{
		{
			name: "plain path",
			pane: Pane{Path: "/test/path"},
			want: "split-window -h -t '=test-session:2' -c '/test/path'",
		},
		{
			name: "path with spaces and quote",
			pane: Pane{Path: "/my path/it's"},
			want: `split-window -h -t '=test-session:2' -c '/my path/it'\''s'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSplitPaneCmd(window, tt.pane)
			if got != tt.want {
				t.Errorf("buildSplitPaneCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestBuildRenamePaneCmd(t *testing.T) {
	window := Window{Index: 2, SessionName: "test-session"}

	tests := []struct {
		name string
		pane Pane
		want string
	}{
		{
			name: "plain title",
			pane: Pane{Title: "test-title"},
			want: "select-pane -t '=test-session:2' -T 'test-title'",
		},
		{
			name: "title with spaces and quote",
			pane: Pane{Title: "it's a title"},
			want: `select-pane -t '=test-session:2' -T 'it'\''s a title'`,
		},
		{
			name: "empty title",
			pane: Pane{},
			want: "select-pane -t '=test-session:2' -T ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRenamePaneCmd(window, tt.pane)
			if got != tt.want {
				t.Errorf("buildRenamePaneCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestBuildNewPanesCmd(t *testing.T) {
	window := Window{Index: 1, SessionName: "test-session"}

	tests := []struct {
		name  string
		panes []Pane
		want  string
	}{
		{
			name:  "no panes",
			panes: nil,
			want:  "",
		},
		{
			name:  "one pane is split then renamed",
			panes: []Pane{{Title: "t1", Path: "/p1"}},
			want: "split-window -h -t '=test-session:1' -c '/p1' ; " +
				"select-pane -t '=test-session:1' -T 't1'",
		},
		{
			name:  "panes keep order",
			panes: []Pane{{Title: "t1", Path: "/p1"}, {Title: "t2", Path: "/p2"}},
			want: "split-window -h -t '=test-session:1' -c '/p1' ; " +
				"select-pane -t '=test-session:1' -T 't1' ; " +
				"split-window -h -t '=test-session:1' -c '/p2' ; " +
				"select-pane -t '=test-session:1' -T 't2'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewPanesCmd(window, tt.panes)
			if got != tt.want {
				t.Errorf("buildNewPanesCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
