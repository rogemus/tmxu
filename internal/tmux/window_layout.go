package tmux

import (
	"encoding/json"
)

type WindowLayout struct {
	Version int              `json:"V"`
	Root    WindowLayoutNode `json:"L"`
}

type WindowLayoutNode struct {
	Type     string             `json:"t"` // "p" pane, "h" side by side, "v" stacked
	Width    int                `json:"w"`
	Height   int                `json:"h"`
	X        int                `json:"x"`
	Y        int                `json:"y"`
	Active   bool               `json:"a"` // only on the active pane
	Index    int                `json:"i"` // pane index, leaves only
	PaneID   string             `json:"I"` // runtime id like "%34", leaves only
	Children []WindowLayoutNode `json:"c"` // splits only
}

// isSplitLayout reports whether the raw layout has more than one pane. An empty
// or invalid layout is not a split.
func isSplitLayout(raw json.RawMessage) bool {
	var wl WindowLayout
	if err := json.Unmarshal(raw, &wl); err != nil {
		return false
	}

	return len(wl.Root.Children) > 0
}

func buildSetLayoutCmd(window Window) []string {
	return []string{"select-layout", "-t", window.TargetIndex, string(window.WindowLayout)}
}
