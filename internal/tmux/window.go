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
	ErrCannotParseLayout   = errors.New("cannot parse window layout")
)

type Window struct {
	ID           string          `json:"-"`
	Slug         string          `json:"-"`
	Index        int             `json:"window_index"`
	Name         string          `json:"window_name"`
	WindowLayout json.RawMessage `json:"window_layout,omitempty"`
	TargetIndex  string          `json:"-"`
}

func (w Window) Layout() (LayoutNode, error) {
	var wl WindowLayout
	if err := json.Unmarshal(w.WindowLayout, &wl); err != nil {
		return LayoutNode{}, fmt.Errorf("%w: %w", ErrCannotParseLayout, err)
	}

	return normalizeWindowLayout(wl), nil
}

func (w Window) Eq(c utils.Comperable) bool {
	i, ok := c.(Window)
	return ok && i.ID == w.ID
}

func CreateWindows(sessionName string, windows []Window, panes map[string][]Pane) error {
	if len(windows) == 0 {
		return nil
	}

	argv, err := buildNewWindowsCmd(sessionName, windows, panes)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateWindows, err)
	}

	out, err := exec.Command("tmux", argv...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s %w", ErrCannotCreateWindows, strings.TrimSpace(string(out)), err)
	}

	return nil
}

func windowSlug(window Window, sessionName string) string {
	return utils.Slug(fmt.Sprintf("%s_%s_%d", sessionName, window.Name, window.Index))
}

func windowTarget(window Window, sessionName string) string {
	return fmt.Sprintf("=%s:%d", sessionName, window.Index)
}

// appendCmd adds cmd to argv, joined by the ";" argument tmux treats as a command separator.
func appendCmd(argv []string, cmd []string) []string {
	if len(argv) > 0 {
		argv = append(argv, ";")
	}

	return append(argv, cmd...)
}

func buildNewWindowsCmd(sessionName string, windows []Window, panesMap map[string][]Pane) ([]string, error) {
	argv := make([]string, 0)

	for _, w := range windows {
		slug := windowSlug(w, sessionName)
		panes, ok := panesMap[slug]
		if !ok || len(panes) == 0 {
			return nil, fmt.Errorf("%w: %s", ErrWindowWithoutPanes, slug)
		}

		firstPane := panes[0]
		restOfPanes := panes[1:]
		argv = appendCmd(argv, buildNewWindowCmd(w, firstPane))
		argv = appendCmd(argv, buildRenamePaneCmd(w, firstPane))

		if len(restOfPanes) > 0 {
			argv = appendCmd(argv, buildNewPanesCmd(w, restOfPanes))
		}

		if len(w.WindowLayout) > 0 {
			argv = appendCmd(argv, buildSetLayoutCmd(w))
		}
	}

	return argv, nil
}

func buildNewWindowCmd(window Window, pane Pane) []string {
	name := utils.Escape(window.Name)
	cmd := []string{"new-window", "-k", "-d", "-t", window.TargetIndex, "-n", name}

	if pane.Path != "" {
		cmd = append(cmd, "-c", pane.Path)
	}

	return cmd
}
