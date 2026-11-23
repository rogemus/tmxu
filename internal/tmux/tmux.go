// Package tmux holds the logic for interacting with the tmux API. It runs tmux
// commands, parses their output, and models sessions, windows and panes.
package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrCannotGetTmuxData  = errors.New("cannot get tmux data")
	ErrCannotLoadSnapshot = errors.New("cannot load snapshot")
	ErrCannotParseLine    = errors.New("cannot parse tmux line")
)

const SchemaVersion = 1

// lineSafeSeparator splits the fields of one `list-panes` row. ASCII Unit
// Separator: tmux names are not expected to contain it, unlike `,` or `:`.
const lineSafeSeparator = "\x1f"

// linePartsCount is the number of fields in one row. Keep it equal to the
// length of `fields` in buildListPanesFormat.
const linePartsCount = 6

type Line struct {
	SessionName  string
	WindowIndex  int
	WindowName   string
	WindowLayout json.RawMessage
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

		windowIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrCannotParseLine, outputLine, err)
		}
		paneIndex, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrCannotParseLine, outputLine, err)
		}

		line := Line{
			SessionName:  parts[0],
			WindowIndex:  windowIndex,
			WindowName:   parts[2],
			WindowLayout: json.RawMessage(parts[3]),
			PaneIndex:    paneIndex,
			PanePath:     parts[5],
		}

		lines = append(lines, line)
	}

	return lines, nil
}

type SnapshotData struct {
	Version  int           `json:"version"`
	Sessions []SessionData `json:"sessions"`
}

func (d SnapshotData) AsTemplate() SnapshotData {
	for si := range d.Sessions {
		s := d.Sessions[si].AsTemplate()
		d.Sessions[si] = s
	}

	return d
}

func (d SnapshotData) ExistingIn(sessions []SessionData) []SessionData {
	existing := make([]SessionData, 0)

	for _, s := range d.Sessions {
		if slices.ContainsFunc(sessions, func(ss SessionData) bool {
			return ss.Name == s.Name
		}) {
			existing = append(existing, s)
		}
	}

	return existing
}

func (d SnapshotData) Template() (SessionData, error) {
	if len(d.Sessions) != 1 {
		return SessionData{}, fmt.Errorf("%w: expected 1 session, got %d", ErrInvalidTemplate, len(d.Sessions))
	}

	return d.Sessions[0], nil
}

func (d SnapshotData) RestoreFromTempl(sessionName, dirPath string) error {
	session, err := d.Template()
	if err != nil {
		return err
	}

	if dirPath != "" {
		session = session.withPaneDirPath(dirPath)
	}

	return RestoreSession(sessionName, dirPath, session.Windows)
}

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

func parseTmuxLines(lines []Line) SnapshotData {
	type winKey struct{ ses, index int }
	sessions := make([]SessionData, 0)

	sessionPos := map[string]int{}
	windowPos := map[winKey]int{}

	for _, l := range lines {
		si, ok := sessionPos[l.SessionName]
		if !ok {
			sessions = append(sessions, SessionData{
				Name:    l.SessionName,
				Windows: make([]WindowData, 0),
			})
			si = len(sessions) - 1
			sessionPos[l.SessionName] = si
		}

		wKey := winKey{si, l.WindowIndex}
		wi, ok := windowPos[wKey]
		if !ok {
			w := WindowData{
				Name:  l.WindowName,
				Index: l.WindowIndex,
				Panes: make([]PaneData, 0),
			}

			if isSplitLayout(l.WindowLayout) {
				w.Layout = l.WindowLayout
			}

			sessions[si].Windows = append(sessions[si].Windows, w)
			wi = len(sessions[si].Windows) - 1
			windowPos[wKey] = wi
		}

		p := PaneData{
			Path:  l.PanePath,
			Index: l.PaneIndex,
		}

		sessions[si].Windows[wi].Panes = append(sessions[si].Windows[wi].Panes, p)
	}

	return SnapshotData{
		Version:  SchemaVersion,
		Sessions: sessions,
	}
}

func buildListPanesFormat() string {
	fields :=
		[]string{
			`#{session_name}`,
			`#{window_index}`,
			`#{window_name}`,
			`#{window_layout}`,
			`#{pane_index}`,
			`#{pane_current_path}`,
		}

	return strings.Join(fields, lineSafeSeparator)
}
