package tmux

import (
	"reflect"
	"testing"
)

// line builds a TmuxLine, deriving names and paths from the ids.
func line(sessionID, windowID, paneID string, windowIndex, paneIndex int) TmuxLine {
	return TmuxLine{
		SessionID:   sessionID,
		SessionName: sessionID + "-name",
		WindowID:    windowID,
		WindowIndex: windowIndex,
		WindowName:  windowID + "-name",
		PaneID:      paneID,
		PaneIndex:   paneIndex,
		PanePath:    "/path/" + paneID,
		PaneTitle:   paneID + "-title",
	}
}

func windowIDs(d TmuxData, sessionID string) []string {
	ids := []string{}
	for _, w := range d.windows[sessionID] {
		ids = append(ids, w.ID)
	}

	return ids
}

func paneIDs(d TmuxData, windowID string) []string {
	ids := []string{}
	for _, p := range d.panes[windowID] {
		ids = append(ids, p.ID)
	}

	return ids
}

func TestParseTmuxLines(t *testing.T) {
	tests := []struct {
		name         string
		lines        []TmuxLine
		wantSessions []string
		wantWindows  map[string][]string // session id -> window ids
		wantPanes    map[string][]string // window id -> pane ids
	}{
		{
			name:         "no lines",
			lines:        nil,
			wantSessions: []string{},
			wantWindows:  map[string][]string{},
			wantPanes:    map[string][]string{},
		},
		{
			name:         "single pane",
			lines:        []TmuxLine{line("session-1", "window-1", "pane-1", 1, 1)},
			wantSessions: []string{"session-1"},
			wantWindows:  map[string][]string{"session-1": {"window-1"}},
			wantPanes:    map[string][]string{"window-1": {"pane-1"}},
		},
		{
			name: "two panes in one window",
			lines: []TmuxLine{
				line("session-1", "window-1", "pane-1", 1, 1),
				line("session-1", "window-1", "pane-2", 1, 2),
			},
			wantSessions: []string{"session-1"},
			wantWindows:  map[string][]string{"session-1": {"window-1"}},
			wantPanes:    map[string][]string{"window-1": {"pane-1", "pane-2"}},
		},
		{
			name: "two windows with two panes each keep both windows",
			lines: []TmuxLine{
				line("session-1", "window-1", "pane-1", 1, 1),
				line("session-1", "window-1", "pane-2", 1, 2),
				line("session-1", "window-2", "pane-3", 2, 1),
				line("session-1", "window-2", "pane-4", 2, 2),
			},
			wantSessions: []string{"session-1"},
			wantWindows:  map[string][]string{"session-1": {"window-1", "window-2"}},
			wantPanes: map[string][]string{
				"window-1": {"pane-1", "pane-2"},
				"window-2": {"pane-3", "pane-4"},
			},
		},
		{
			name: "three windows in a row keep all windows",
			lines: []TmuxLine{
				line("session-1", "window-1", "pane-1", 1, 1),
				line("session-1", "window-2", "pane-2", 2, 1),
				line("session-1", "window-3", "pane-3", 3, 1),
				line("session-1", "window-3", "pane-4", 3, 2),
			},
			wantSessions: []string{"session-1"},
			wantWindows:  map[string][]string{"session-1": {"window-1", "window-2", "window-3"}},
			wantPanes: map[string][]string{
				"window-1": {"pane-1"},
				"window-2": {"pane-2"},
				"window-3": {"pane-3", "pane-4"},
			},
		},
		{
			name: "two sessions keep their own windows and panes",
			lines: []TmuxLine{
				line("session-1", "window-1", "pane-1", 1, 1),
				line("session-1", "window-1", "pane-2", 1, 2),
				line("session-2", "window-2", "pane-3", 1, 1),
				line("session-2", "window-3", "pane-4", 2, 1),
			},
			wantSessions: []string{"session-1", "session-2"},
			wantWindows: map[string][]string{
				"session-1": {"window-1"},
				"session-2": {"window-2", "window-3"},
			},
			wantPanes: map[string][]string{
				"window-1": {"pane-1", "pane-2"},
				"window-2": {"pane-3"},
				"window-3": {"pane-4"},
			},
		},
		{
			name: "duplicate line is not added twice",
			lines: []TmuxLine{
				line("session-1", "window-1", "pane-1", 1, 1),
				line("session-1", "window-1", "pane-1", 1, 1),
			},
			wantSessions: []string{"session-1"},
			wantWindows:  map[string][]string{"session-1": {"window-1"}},
			wantPanes:    map[string][]string{"window-1": {"pane-1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTmuxLines(tt.lines)

			gotSessions := []string{}
			for id := range got.sessions {
				gotSessions = append(gotSessions, id)
			}
			if len(gotSessions) != len(tt.wantSessions) {
				t.Fatalf("sessions = %v, want %v", gotSessions, tt.wantSessions)
			}
			for _, id := range tt.wantSessions {
				if _, ok := got.sessions[id]; !ok {
					t.Errorf("session %s missing, got %v", id, gotSessions)
				}
			}

			if len(got.windows) != len(tt.wantWindows) {
				t.Errorf("windows keys = %d, want %d", len(got.windows), len(tt.wantWindows))
			}
			for sid, want := range tt.wantWindows {
				if g := windowIDs(got, sid); !reflect.DeepEqual(g, want) {
					t.Errorf("windows[%s] = %v, want %v", sid, g, want)
				}
			}

			if len(got.panes) != len(tt.wantPanes) {
				t.Errorf("panes keys = %d, want %d", len(got.panes), len(tt.wantPanes))
			}
			for wid, want := range tt.wantPanes {
				if g := paneIDs(got, wid); !reflect.DeepEqual(g, want) {
					t.Errorf("panes[%s] = %v, want %v", wid, g, want)
				}
			}
		})
	}
}

func TestParseTmuxLinesFields(t *testing.T) {
	got := parseTmuxLines([]TmuxLine{line("test-session-id", "test-window-id", "test-pane-id", 4, 2)})

	s := got.sessions["test-session-id"]
	if s.ID != "test-session-id" || s.Name != "test-session-id-name" {
		t.Errorf("session = %+v", s)
	}

	w := got.windows["test-session-id"][0]
	if w.ID != "test-window-id" || w.Index != 4 || w.Name != "test-window-id-name" {
		t.Errorf("window = %+v", w)
	}

	p := got.panes["test-window-id"][0]
	if p.ID != "test-pane-id" || p.Index != 2 || p.Path != "/path/test-pane-id" || p.Title != "test-pane-id-title" {
		t.Errorf("pane = %+v", p)
	}
}
