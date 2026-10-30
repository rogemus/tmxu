package tmux

type Window struct {
	Id    string `json:"id"`
	Index int    `json:"index"`
	Name  string `json:"name"`
	Panes []Pane `json:"panes"`
}

type WindowLayout struct {
}

func NewWindow(l TmuxLine) Window {
	return Window{
		Id:    l.WindowID,
		Name:  l.WindowName,
		Index: l.WindowIndex,
		Panes: make([]Pane, 0),
	}
}

func (w Window) Eq(c Comperable) bool {
	i, ok := c.(Window)
	return ok && i.Id == w.Id
}
