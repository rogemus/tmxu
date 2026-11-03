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
	TargetIndex  string          `json:"window_target"`
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

func CreateWindows(windows []Window, panes map[string][]Pane) error {
	if len(windows) == 0 {
		return nil
	}

	cmd := buildNewWindowsCmd(windows, panes)
	_, err := exec.Command("tmux", cmd).Output()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateWindows, err)
	}

	return nil
}

func buildNewWindowsCmd(windows []Window, panesMap map[string][]Pane) string {
	cmds := make([]string, 0)

	for _, w := range windows {
		panes, ok := panesMap[w.ID]
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

	cmd := []string{"new-window", "-k", "-d", "-t", target, "-n", name, "-c", utils.Quote(pane.Path)}
	return strings.Join(cmd, " ")
}
