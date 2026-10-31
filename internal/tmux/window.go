package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateWindows = errors.New("cannot create windows")
)

type Window struct {
	ID          string `json:"id"`
	Index       int    `json:"index"`
	Name        string `json:"name"`
	SessionName string `json:"sessionName"`
}

func NewWindow(l Line) Window {
	return Window{
		ID:          l.WindowID,
		Name:        l.WindowName,
		Index:       l.WindowIndex,
		SessionName: l.SessionName,
	}
}

func (w Window) GetName() string { return utils.Quote(w.Name) }

func (w Window) SessionIndex() string {
	return utils.Quote(fmt.Sprintf("=%s:%d", w.SessionName, w.Index))
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
	}

	return strings.Join(cmds, " ; ")
}

func buildNewWindowCmd(window Window, pane Pane) string {
	cmd := []string{"new-window", "-k", "-d", "-t", window.SessionIndex(), "-n", window.GetName(), "-c", pane.GetPath()}
	return strings.Join(cmd, " ")
}
