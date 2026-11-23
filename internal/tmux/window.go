package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateWindows = errors.New("cannot create windows")
	ErrWindowWithoutPanes  = errors.New("window has no panes")
)

type WindowData struct {
	Name   string          `json:"window_name"`
	Index  int             `json:"window_index"`
	Layout json.RawMessage `json:"window_layout,omitempty"`
	Panes  []PaneData      `json:"window_panes,omitempty"`
}

func CreateWindows(sessionName string, windows []WindowData) error {
	if len(windows) == 0 {
		return nil
	}

	argv, err := buildNewWindowsCmd(sessionName, windows)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateWindows, err)
	}

	out, err := exec.Command("tmux", argv...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrCannotCreateWindows, strings.TrimSpace(string(out)), err)
	}

	return nil
}

func windowTarget(sessionName string, winIndex int) string {
	return fmt.Sprintf("=%s:%d", sessionName, winIndex)
}

// appendCmd adds cmd to argv, joined by the ";" argument tmux treats as a command separator.
func appendCmd(argv []string, cmd []string) []string {
	if len(argv) > 0 {
		argv = append(argv, ";")
	}

	return append(argv, cmd...)
}

func buildNewWindowsCmd(sessionName string, windows []WindowData) ([]string, error) {
	argv := make([]string, 0)

	for _, w := range windows {
		panes := w.Panes

		if len(panes) == 0 {
			return nil, fmt.Errorf("%w: %s", ErrWindowWithoutPanes, w.Name)
		}

		firstPane := panes[0]
		restOfPanes := panes[1:]
		winTarget := windowTarget(sessionName, w.Index)

		argv = appendCmd(argv, buildNewFirstWindowCmd(winTarget, w.Name, firstPane))
		argv = appendCmd(argv, buildRenamePaneCmd(winTarget, w.Name, firstPane))

		if len(restOfPanes) > 0 {
			argv = appendCmd(argv, buildNewPanesCmd(winTarget, w.Name, restOfPanes))
		}

		if len(w.Layout) > 0 {
			argv = appendCmd(argv, buildSetLayoutCmd(winTarget, w.Layout))
		}
	}

	return argv, nil
}

// buildNewFirstWindowCmd builds the command that recreates the first window of
// a restored session. -k replaces the window new-session already created at
// that index, so it is needed only here.
func buildNewFirstWindowCmd(winTarget, winName string, pane PaneData) []string {
	name := utils.Escape(winName)
	cmd := []string{"new-window", "-k", "-d", "-t", winTarget, "-n", name}

	if pane.Path != "" {
		cmd = append(cmd, "-c", pane.Path)
	}

	return cmd
}
