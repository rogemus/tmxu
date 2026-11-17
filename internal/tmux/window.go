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

	cmd := buildNewWindowsCmd(sessionName, windows, panes)
	out, err := exec.Command("tmux", cmd).CombinedOutput()
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

func buildNewWindowsCmd(sessionName string, windows []Window, panesMap map[string][]Pane) string {
	cmds := make([]string, 0)

	for _, w := range windows {
		slug := windowSlug(w, sessionName)
		panes, ok := panesMap[slug]
		if !ok {
			panic("windows need to always have at least 1 pane")
		}

		firstPane := panes[0]
		restOfPanes := panes[1:]
		cmds = append(cmds, buildNewWindowCmd(w, firstPane))
		cmds = append(cmds, buildRenamePaneCmd(w, firstPane))

		if len(restOfPanes) > 0 {
			cmds = append(cmds, buildNewPanesCmd(w, restOfPanes))
		}

		if w.WindowLayout != nil {
			cmds = append(cmds, buildSetLayoutCmd(w))
		}
	}

	return strings.Join(cmds, " ; ")
}

func buildNewWindowCmd(window Window, pane Pane) string {
	target := utils.Quote(window.TargetIndex)
	name := utils.Quote(utils.Escape(window.Name))
	cmd := []string{"new-window", "-k", "-d", "-t", target, "-n", name}

	if pane.Path != "" {
		cmd = append(cmd, "-c", utils.Quote(pane.Path))
	}

	return strings.Join(cmd, " ")
}
