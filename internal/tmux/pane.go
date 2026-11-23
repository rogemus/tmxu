package tmux

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/utils"
)

type PaneData struct {
	Index int    `json:"pane_index"`
	Path  string `json:"pane_path"`
}

func (d *PaneData) AsTemplate() *PaneData {
	d.Path = ""
	return d
}

func paneTitle(windowName string, index int) string {
	return fmt.Sprintf("%s_%d", windowName, index)
}

func buildNewPanesCmd(winTarget, winName string, panes []PaneData) []string {
	argv := make([]string, 0)

	for _, p := range panes {
		argv = appendCmd(argv, buildSplitPaneCmd(winTarget, p))
		argv = appendCmd(argv, buildRenamePaneCmd(winTarget, winName, p))
	}

	return argv
}

func buildSplitPaneCmd(winTarget string, pane PaneData) []string {
	cmd := []string{"split-window", "-h", "-t", winTarget}

	if pane.Path != "" {
		cmd = append(cmd, "-c", pane.Path)
	}

	return cmd
}

func buildRenamePaneCmd(winTarget, winName string, pane PaneData) []string {
	paneTarget := fmt.Sprintf("%s.%d", winTarget, pane.Index)
	title := utils.Escape(paneTitle(winName, pane.Index))

	return []string{"select-pane", "-t", paneTarget, "-T", title}
}
