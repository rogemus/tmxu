package tmux

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/utils"
)

type Pane struct {
	ID    string `json:"-"`
	Index int    `json:"pane_index"`
	Title string `json:"pane_title"`
	Path  string `json:"pane_path"`
}

func buildNewPanesCmd(window Window, panes []Pane) []string {
	argv := make([]string, 0)
	for _, p := range panes {
		argv = appendCmd(argv, buildSplitPaneCmd(window, p))
		argv = appendCmd(argv, buildRenamePaneCmd(window, p))
	}

	return argv
}

func buildSplitPaneCmd(window Window, pane Pane) []string {
	cmd := []string{"split-window", "-h", "-t", window.TargetIndex}

	if pane.Path != "" {
		cmd = append(cmd, "-c", pane.Path)
	}

	return cmd
}

func buildRenamePaneCmd(window Window, pane Pane) []string {
	paneTarget := fmt.Sprintf("%s.%d", window.TargetIndex, pane.Index)
	title := utils.Escape(pane.Title)

	return []string{"select-pane", "-t", paneTarget, "-T", title}
}
