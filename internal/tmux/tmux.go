// Package tmux holds the logic for interacting with the tmux API. It runs tmux
// commands, parses their output, and models sessions, windows and panes.
package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotGetTmuxData        = errors.New("cannot get tmux data")
	ErrCannotLoadSnapshot       = errors.New("cannot load snapshot")
	ErrCannotSanitizeSnapshot   = errors.New("cannot sanitize snapshot")
	ErrCannotDesanitizeSnapshot = errors.New("cannot desanitize snapshot")
	ErrCannotParseLine          = errors.New("cannot parse tmux line")
)

const SchemaVersion = 2

// templateSessionName is the session name stored in a template, which has no
// session of its own until a new one is created from it.
const templateSessionName = ""

// lineSafeSeparator splits the fields of one `list-panes` row. ASCII Unit
// Separator: tmux names are not expected to contain it, unlike `,` or `:`.
const lineSafeSeparator = "\x1f"

// linePartsCount is the number of fields in one row. Keep it equal to the
// length of `fields` in buildListPanesFormat.
const linePartsCount = 9

type Line struct {
	SessionID    string
	SessionName  string
	WindowID     string
	WindowIndex  int
	WindowName   string
	WindowLayout json.RawMessage
	PaneID       string
	PaneIndex    int
	PanePath     string
}

func parseLines(output []byte) ([]Line, error) {
	lines := make([]Line, 0)
	outputLines := strings.SplitSeq(strings.TrimSuffix(string(output), "\n"), "\n")

	for outputLine := range outputLines {
		parts := strings.SplitN(outputLine, lineSafeSeparator, linePartsCount)
		if len(parts) != linePartsCount {
			return nil, fmt.Errorf("%w: missing parts %d out of %d", ErrCannotParseLine, len(parts), linePartsCount)
		}

		windowIndex, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrCannotParseLine, outputLine, err)
		}
		paneIndex, err := strconv.Atoi(parts[7])
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrCannotParseLine, outputLine, err)
		}

		line := Line{
			SessionID:    parts[0],
			SessionName:  parts[1],
			WindowID:     parts[2],
			WindowIndex:  windowIndex,
			WindowName:   parts[4],
			WindowLayout: json.RawMessage(parts[5]),
			PaneID:       parts[6],
			PaneIndex:    paneIndex,
			PanePath:     parts[8],
		}

		lines = append(lines, line)
	}

	return lines, nil
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

// Snapshot reads the live tmux state. An empty sessionName snapshots every
// session on the server, otherwise only that session, or ErrNoSession if it
// does not exist.
func Snapshot(sessionName string) (SnapshotData, error) {
	argv := []string{
		"list-panes", "-F", buildListPanesFormat(),
	}

	if sessionName != "" {
		hasSession, err := HasSession(sessionName)
		if err != nil {
			return SnapshotData{}, err
		}

		if !hasSession {
			return SnapshotData{}, fmt.Errorf("%w: %s", ErrNoSession, sessionName)
		}

		argv = append(argv, "-s", "-t", "="+sessionName)
	} else {
		argv = append(argv, "-a")
	}

	o, err := exec.Command("tmux", argv...).Output()
	if err != nil {
		return SnapshotData{}, fmt.Errorf("%w: %w", ErrCannotGetTmuxData, err)
	}

	lines, err := parseLines(o)
	if err != nil {
		return SnapshotData{}, err
	}

	data := parseTmuxLines(lines)
	return data, nil
}

func LoadSnapshot(data []byte) (SnapshotData, error) {
	snapshot := SnapshotData{}

	err := json.Unmarshal(data, &snapshot)
	if err != nil {
		return SnapshotData{}, fmt.Errorf("%w: %w", ErrCannotLoadSnapshot, err)
	}

	return snapshot, nil
}

func DesanitizeTemplSnapshot(data SnapshotData, sessionName string) (SnapshotData, error) {
	oldSSlug := sessionSlug(templateSessionName)
	session, ok := data.Sessions[oldSSlug]
	if !ok {
		return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotDesanitizeSnapshot, sessionName)
	}

	sSlug := sessionSlug(sessionName)
	session.Slug = sSlug
	session.Name = sessionName

	data.Sessions[sSlug] = session
	delete(data.Sessions, oldSSlug)

	windows, ok := data.Windows[oldSSlug]
	if !ok {
		return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotDesanitizeSnapshot, sessionName)
	}

	for i := range windows {
		w := &windows[i]
		wSlug := windowSlug(*w, sSlug)
		oldWSlug := windowSlug(*w, templateSessionName)
		target := windowTarget(*w, sessionName)

		w.Slug = wSlug
		w.TargetIndex = target

		panes, ok := data.Panes[oldWSlug]
		if !ok {
			return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotDesanitizeSnapshot, sessionName)
		}

		data.Panes[wSlug] = panes
		delete(data.Panes, oldWSlug)
	}

	data.Windows[sSlug] = windows
	delete(data.Windows, oldSSlug)

	return data, nil
}

func SanitizeSnapshot(data SnapshotData, sessionName string) (SnapshotData, error) {
	oldSSlug := sessionSlug(sessionName)
	session, ok := data.Sessions[oldSSlug]
	if !ok {
		return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotSanitizeSnapshot, sessionName)
	}

	sSlug := sessionSlug(templateSessionName)
	session.Slug = sSlug
	session.Name = templateSessionName

	data.Sessions[sSlug] = session
	delete(data.Sessions, oldSSlug)

	windows, ok := data.Windows[oldSSlug]
	if !ok {
		return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotSanitizeSnapshot, sessionName)
	}

	for i := range windows {
		w := &windows[i]
		wSlug := windowSlug(*w, templateSessionName)
		target := windowTarget(*w, templateSessionName)

		oldWSlug := w.Slug
		w.Slug = wSlug
		w.TargetIndex = target

		panes, ok := data.Panes[oldWSlug]
		if !ok {
			return SnapshotData{}, fmt.Errorf("%w: %q", ErrCannotSanitizeSnapshot, sessionName)
		}

		for j := range panes {
			p := &panes[j]
			p.Path = ""
		}

		data.Panes[wSlug] = panes
		delete(data.Panes, oldWSlug)
	}

	data.Windows[sSlug] = windows
	delete(data.Windows, oldSSlug)

	return data, nil
}

func parseTmuxLines(lines []Line) SnapshotData {
	sessions := make(map[string]Session)
	windows := make(map[string][]Window)
	panes := make(map[string][]Pane)

	for _, l := range lines {
		s := l.Session()
		if existingS, ok := sessions[s.Slug]; ok {
			s = existingS
		}
		sessions[s.Slug] = s

		w := l.Window()
		if !slices.ContainsFunc(windows[s.Slug], func(win Window) bool { return win.ID == w.ID }) {
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

// buildListPanesFormat returns the `list-panes -F` format: one row per pane,
// fields joined by lineSafeSeparator. Field order must match parseLines.
// pane_current_path stays last so SplitN keeps any stray separator in it.
func buildListPanesFormat() string {
	fields :=
		[]string{
			`#{session_id}`,
			`#{session_name}`,
			`#{window_id}`,
			`#{window_index}`,
			`#{window_name}`,
			`#{window_layout}`,
			`#{pane_id}`,
			`#{pane_index}`,
			`#{pane_current_path}`,
		}

	return strings.Join(fields, lineSafeSeparator)
}
