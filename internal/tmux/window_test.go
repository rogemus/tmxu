package tmux

import (
	"strings"
	"testing"
)

func TestWindowSessionIndex(t *testing.T) {
	w := Window{Index: 3, SessionName: "test-session"}

	got := w.SessionIndex()
	if want := "'test-session:3'"; got != want {
		t.Errorf("SessionIndex() = %s, want %s", got, want)
	}
}

func TestBuildNewWindowCmd(t *testing.T) {
	tests := []struct {
		name   string
		window Window
		pane   Pane
		want   string
	}{
		{
			name:   "plain values",
			window: Window{Index: 1, Name: "test-window", SessionName: "test-session"},
			pane:   Pane{Path: "/test/path"},
			want:   "new-window -k -d -t 'test-session:1' -n 'test-window' -c '/test/path'",
		},
		{
			name:   "index is part of the target",
			window: Window{Index: 4, Name: "test-window", SessionName: "test-session"},
			pane:   Pane{Path: "/test/path"},
			want:   "new-window -k -d -t 'test-session:4' -n 'test-window' -c '/test/path'",
		},
		{
			name:   "spaces in session, window and path",
			window: Window{Index: 2, Name: "test window", SessionName: "test session"},
			pane:   Pane{Path: "/test/my path"},
			want:   "new-window -k -d -t 'test session:2' -n 'test window' -c '/test/my path'",
		},
		{
			name:   "single quote in window name",
			window: Window{Index: 1, Name: "it's", SessionName: "test-session"},
			pane:   Pane{Path: "/test/path"},
			want:   `new-window -k -d -t 'test-session:1' -n 'it'\''s' -c '/test/path'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewWindowCmd(tt.window, tt.pane)
			if got != tt.want {
				t.Errorf("buildNewWindowCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestBuildNewWindowsCmd(t *testing.T) {
	window1 := Window{ID: "test-window-1", Index: 1, Name: "test-window-1", SessionName: "test-session"}
	window2 := Window{ID: "test-window-2", Index: 2, Name: "test-window-2", SessionName: "test-session"}

	pane := func(id string) Pane {
		return Pane{ID: id, Title: id + "-title", Path: "/test/" + id}
	}

	tests := []struct {
		name     string
		windows  []Window
		panesMap map[string][]Pane
		want     string
	}{
		{
			name:     "no windows",
			windows:  nil,
			panesMap: map[string][]Pane{},
			want:     "",
		},
		{
			name:    "one window with two panes",
			windows: []Window{window1},
			panesMap: map[string][]Pane{
				"test-window-1": {pane("pane-1"), pane("pane-2")},
			},
			want: strings.Join([]string{
				"new-window -k -d -t 'test-session:1' -n 'test-window-1' -c '/test/pane-1'",
				"select-pane -t 'test-session:1' -T 'pane-1-title'",
				"split-window -h -t 'test-session:1' -c '/test/pane-2'",
				"select-pane -t 'test-session:1' -T 'pane-2-title'",
			}, " ; "),
		},
		{
			name:    "two windows keep their own panes and order",
			windows: []Window{window1, window2},
			panesMap: map[string][]Pane{
				"test-window-1": {pane("pane-1"), pane("pane-2")},
				"test-window-2": {pane("pane-3"), pane("pane-4")},
			},
			want: strings.Join([]string{
				"new-window -k -d -t 'test-session:1' -n 'test-window-1' -c '/test/pane-1'",
				"select-pane -t 'test-session:1' -T 'pane-1-title'",
				"split-window -h -t 'test-session:1' -c '/test/pane-2'",
				"select-pane -t 'test-session:1' -T 'pane-2-title'",
				"new-window -k -d -t 'test-session:2' -n 'test-window-2' -c '/test/pane-3'",
				"select-pane -t 'test-session:2' -T 'pane-3-title'",
				"split-window -h -t 'test-session:2' -c '/test/pane-4'",
				"select-pane -t 'test-session:2' -T 'pane-4-title'",
			}, " ; "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNewWindowsCmd(tt.windows, tt.panesMap)
			if got != tt.want {
				t.Errorf("buildNewWindowsCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestBuildNewWindowsCmdPanicsWithoutPanes(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected a panic for a window without panes")
		}
	}()

	buildNewWindowsCmd([]Window{{ID: "test-window-1"}}, map[string][]Pane{})
}

func TestBuildNewWindowsCmdWindowCommandComesFirst(t *testing.T) {
	window := Window{ID: "test-window-1", Index: 1, Name: "test-window-1", SessionName: "test-session"}
	panesMap := map[string][]Pane{
		"test-window-1": {
			{ID: "pane-1", Title: "pane-1-title", Path: "/test/pane-1"},
			{ID: "pane-2", Title: "pane-2-title", Path: "/test/pane-2"},
		},
	}

	got := buildNewWindowsCmd([]Window{window}, panesMap)

	newWindow := strings.Index(got, "new-window")
	firstSelect := strings.Index(got, "select-pane")
	split := strings.Index(got, "split-window")
	if newWindow != 0 || firstSelect < newWindow || split < firstSelect {
		t.Errorf("commands out of order, new-window=%d select-pane=%d split-window=%d in\n%s", newWindow, firstSelect, split, got)
	}
}

// A window with a single pane has no extra panes to split, so the chain must
// not contain an empty command (which would leave "a ;  ; b" or a trailing ";").
func TestBuildNewWindowsCmdNoEmptyCommands(t *testing.T) {
	window1 := Window{ID: "test-window-1", Index: 1, Name: "test-window-1", SessionName: "test-session"}
	window2 := Window{ID: "test-window-2", Index: 2, Name: "test-window-2", SessionName: "test-session"}
	panesMap := map[string][]Pane{
		"test-window-1": {{ID: "pane-1", Title: "pane-1-title", Path: "/test/pane-1"}},
		"test-window-2": {{ID: "pane-2", Title: "pane-2-title", Path: "/test/pane-2"}},
	}

	got := buildNewWindowsCmd([]Window{window1, window2}, panesMap)

	for i, cmd := range strings.Split(got, " ; ") {
		if strings.TrimSpace(cmd) == "" {
			t.Errorf("command %d is empty in\n%q", i, got)
		}
	}
}
