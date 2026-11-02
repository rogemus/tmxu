package tmux

import (
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

type Pane struct {
	ID    string `json:"pane_id"`
	Index int    `json:"pane_index"`
	Title string `json:"pane_title"`
	Path  string `json:"pane_path"`
}

func (p Pane) GetTitle() string { return utils.Quote(p.Title) }

func (p Pane) GetPath() string { return utils.Quote(p.Path) }

func (p Pane) Eq(c utils.Comperable) bool {
	i, ok := c.(Pane)
	return ok && p.ID == i.ID
}

func buildNewPanesCmd(window Window, panes []Pane) string {
	cmds := make([]string, 0)
	for _, p := range panes {
		cmds = append(cmds, buildSplitPaneCmd(window, p))
		cmds = append(cmds, buildRenamePaneCmd(window, p))
	}

	return strings.Join(cmds, " ; ")
}

func buildSplitPaneCmd(window Window, pane Pane) string {
	cmd := []string{"split-window", "-h", "-t", window.SessionIndex(), "-c", pane.GetPath()}
	return strings.Join(cmd, " ")
}

func buildRenamePaneCmd(window Window, pane Pane) string {
	cmd := []string{"select-pane", "-t", window.SessionIndex(), "-T", pane.GetTitle()}
	return strings.Join(cmd, " ")
}
