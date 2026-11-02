package tmux

import (
	"fmt"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

type Pane struct {
	ID    string `json:"pane_id"`
	Index int    `json:"pane_index"`
	Title string `json:"pane_title"`
	Path  string `json:"pane_path"`
}

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
	target := utils.Quote(window.TargetIndex())
	cmd := []string{"split-window", "-h", "-t", target, "-c", utils.Quote(pane.Path)}
	return strings.Join(cmd, " ")
}

func buildRenamePaneCmd(window Window, pane Pane) string {
	target := window.TargetIndex()
	paneTarget := fmt.Sprintf("%s.%d", target, pane.Index)
	title := utils.Quote(utils.Escape(pane.Title))

	cmd := []string{"select-pane", "-t", utils.Quote(paneTarget), "-T", title}
	return strings.Join(cmd, " ")
}
