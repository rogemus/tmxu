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

type TmuxLine struct {
	SessionID   string `json:"session_id"`
	SessionName string `json:"session_name"`
	WindowID    string `json:"window_id"`
	WindowIndex int    `json:"window_index"`
	WindowName  string `json:"window_name"`
	PaneID      string `json:"pane_id"`
	PaneIndex   int    `json:"pane_index"`
	PanePath    string `json:"pane_path"`
	PaneTitle   string `json:"pane_title"`
}

type TmuxData struct {
	sessions map[string]Session
	windows  map[string][]Window
	panes    map[string][]Pane
}

func (d TmuxData) String() string {
	ids := make([]string, 0, len(d.sessions))
	for id := range d.sessions {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	fmt.Fprintf(&b, "TmuxData: %d sessions\n", len(ids))

	for _, sid := range ids {
		fmt.Fprintf(&b, "%s %q\n", sid, d.sessions[sid].Name)

		for _, w := range d.windows[sid] {
			fmt.Fprintf(&b, "  %s [%d] %q\n", w.ID, w.Index, w.Name)

			for _, p := range d.panes[w.ID] {
				fmt.Fprintf(&b, "    %s [%d] %q %s\n", p.ID, p.Index, p.Title, p.Path)
			}
		}
	}

	return b.String()
}

func GetTmuxData() (TmuxData, error) {
	format := "{" + strings.Join(
		[]string{
			`"session_id": "#{session_id}"`,
			`"session_name": "#{session_name}"`,
			`"window_id": "#{window_id}"`,
			`"window_index": #{window_index}`,
			`"window_name": "#{window_name}"`,
			`"pane_id": "#{pane_id}"`,
			`"pane_index": #{pane_index}`,
			`"pane_path": "#{pane_current_path}"`,
			`"pane_title": "#{pane_title}"`,
		}, ",",
	) + "},"

	o, err := exec.Command("tmux", "list-panes", "-a", "-F", format).Output()
	if err != nil {
		return TmuxData{}, fmt.Errorf("%w: %w", ErrCannotGetTmuxData, err)
	}

	doc := "[" + strings.TrimSuffix(strings.TrimSpace(string(o)), ",") + "]"
	var l []TmuxLine
	err = json.Unmarshal([]byte(doc), &l)
	if err != nil {
		return TmuxData{}, fmt.Errorf("%w: %w", ErrCannotParseTmuxLines, err)
	}

	data := parseTmuxLines(l)
	return data, nil
}

func parseTmuxLines(lines []TmuxLine) TmuxData {
	sessions := make(map[string]Session, 0)
	windows := make(map[string][]Window, 0)
	panes := make(map[string][]Pane, 0)

	for _, l := range lines {

		s := NewSession(l)
		if _, ok := sessions[l.SessionID]; ok {
			s = sessions[l.SessionID]
		} else {
			sessions[s.ID] = s
		}

		w := NewWindow(l)
		if !utils.Contain(windows[s.ID], w) {
			windows[s.ID] = append(windows[s.ID], w)
		}

		p := NewPane(l)
		if !utils.Contain(panes[w.ID], p) {
			panes[w.ID] = append(panes[w.ID], p)
		}
	}

	return TmuxData{sessions, windows, panes}
}
