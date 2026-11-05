// Package tmux holds the logic for interacting with the tmux API. It runs tmux
// commands, parses their output, and models sessions, windows and panes.
package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotGetTmuxData    = errors.New("cannot get tmux data")
	ErrCannotParseTmuxLines = errors.New("cannot parse tmux lines")
)

const SchemaVersion = 2

type Line struct {
	SessionID    string          `json:"session_id"`
	SessionName  string          `json:"session_name"`
	WindowID     string          `json:"window_id"`
	WindowIndex  int             `json:"window_index"`
	WindowName   string          `json:"window_name"`
	WindowLayout json.RawMessage `json:"window_layout"`
	PaneID       string          `json:"pane_id"`
	PaneIndex    int             `json:"pane_index"`
	PanePath     string          `json:"pane_path"`
}

func (l Line) Session() Session {
	return Session{
		ID:   l.SessionID,
		Slug: utils.Slug(l.SessionName),
		Name: l.SessionName,
	}
}

func (l Line) Window() Window {
	slug := utils.Slug(fmt.Sprintf("%s_%s_%d", l.SessionName, l.WindowName, l.WindowIndex))
	target := fmt.Sprintf("=%s:%d", l.SessionName, l.WindowIndex)

	w := Window{
		ID:          l.WindowID,
		Slug:        slug,
		Name:        l.WindowName,
		Index:       l.WindowIndex,
		TargetIndex: target,
	}

	// A single pane fills the window, so only split windows keep a layout.
	if isSplitLayout(l.WindowLayout) {
		w.WindowLayout = l.WindowLayout
	}

	return w
}

func (l Line) Pane() Pane {
	return Pane{
		ID:    l.PaneID,
		Index: l.PaneIndex,
		Path:  l.PanePath,
		Title: fmt.Sprintf("%s_%d", l.WindowName, l.PaneIndex),
	}
}

type SnapshotData struct {
	Version  int                 `json:"version"`
	Sessions map[string]Session  `json:"sessions"`
	Windows  map[string][]Window `json:"windows"`
	Panes    map[string][]Pane   `json:"panes"`
}

func (d SnapshotData) sessionIDs() []string {
	ids := make([]string, 0, len(d.Sessions))
	for id := range d.Sessions {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	return ids
}

func Snapshot() (SnapshotData, error) {
	o, err := exec.Command("tmux", "list-panes", "-a", "-F", buildListPanesFormat()).Output()
	if err != nil {
		return SnapshotData{}, fmt.Errorf("%w: %w", ErrCannotGetTmuxData, err)
	}

	var l []Line
	err = json.Unmarshal(normalizeListPanesOut(o), &l)
	if err != nil {
		return SnapshotData{}, fmt.Errorf("%w: %w", ErrCannotParseTmuxLines, err)
	}

	data := parseTmuxLines(l)
	return data, nil
}

func parseTmuxLines(lines []Line) SnapshotData {
	sessions := make(map[string]Session, 0)
	windows := make(map[string][]Window, 0)
	panes := make(map[string][]Pane, 0)

	for _, l := range lines {
		s := l.Session()
		if _, ok := sessions[s.Slug]; ok {
			s = sessions[s.Slug]
		} else {
			sessions[s.Slug] = s
		}

		w := l.Window()
		if !utils.Contain(windows[s.Slug], w) {
			windows[s.Slug] = append(windows[s.Slug], w)
		}

		p := l.Pane()
		panes[w.Slug] = append(panes[w.Slug], p)
	}

	return SnapshotData{
		Version:  SchemaVersion,
		Sessions: sessions,
		Windows:  windows,
		Panes:    panes,
	}
}

func buildListPanesFormat() string {
	fields :=
		[]string{
			`"session_id": "#{session_id}"`,
			`"session_name": "#{session_name}"`,
			`"window_id": "#{window_id}"`,
			`"window_index": #{window_index}`,
			`"window_name": "#{window_name}"`,
			`"window_layout": #{window_layout}`,
			`"pane_id": "#{pane_id}"`,
			`"pane_index": #{pane_index}`,
			`"pane_path": "#{pane_current_path}"`,
		}

	return fmt.Sprintf("{%s},", strings.Join(fields, ","))
}

func normalizeListPanesOut(output []byte) []byte {
	strOut := strings.TrimSpace(string(output))
	cleaned := strings.TrimSuffix(strOut, ",")

	return []byte(fmt.Sprintf("[%s]", cleaned))
}
