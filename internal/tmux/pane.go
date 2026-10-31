package tmux

import (
	"errors"
	"strings"
)

type Pane struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

var (
	ErrCannotCreatePane = errors.New("cannot create pane")
	ErrCannotRenamePane = errors.New("cannot rename pane")
	ErrCannotListPanes  = errors.New("cannot list panes")
)

func NewPane(l TmuxLine) Pane {
	return Pane{
		ID:    l.PaneID,
		Index: l.PaneIndex,
		Title: l.PaneTitle,
		Path:  l.PanePath,
	}
}

func (p Pane) StrTitle() string { return "\"" + p.Title + "\"" }

func (p Pane) StrPath() string { return "\"" + p.Path + "\"" }

func (p Pane) Eq(c Comperable) bool {
	i, ok := c.(Pane)
	return ok && p.ID == i.ID
}

func buildNewPanesCmd(window Window, panes []Pane) string {
	if len(panes) == 0 {
		return ""
	}

	cmds := make([]string, 0)
	for _, p := range panes {
		cmds = append(cmds, buildSplitPaneCmd(window, p))
		cmds = append(cmds, buildRenamePaneCmd(window, p))
	}

	return strings.Join(cmds, " ; ")
}

func buildSplitPaneCmd(window Window, pane Pane) string {
	cmd := []string{"split-window", "-h", "-t", window.SessionIndex(), "-c", pane.StrPath()}
	return strings.Join(cmd, " ")
}

func buildRenamePaneCmd(window Window, pane Pane) string {
	cmd := []string{"select-pane", "-t", window.SessionIndex(), "-T", pane.StrTitle()}
	return strings.Join(cmd, " ")
}
