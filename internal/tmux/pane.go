package tmux

import (
	"errors"
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

func (p Pane) Eq(c Comperable) bool {
	i, ok := c.(Pane)
	return ok && p.ID == i.ID
}
