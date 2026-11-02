package tmux

import (
	"encoding/json"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
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

const (
	SplitHorizontal = "horizontal" // panes side by side
	SplitVertical   = "vertical"   // panes stacked
)

type LayoutNode struct {
	Split string       `json:"split,omitempty"` // empty for a pane
	Size  int          `json:"size"`            // percent of the parent
	Panes []LayoutNode `json:"panes,omitempty"` // splits only
}

func normalizeWindowLayout(l WindowLayout) LayoutNode {
	root := normalizeLayoutNode(l.Root, nil)
	root.Size = 100

	return root
}

func normalizeSplit(t string) string {
	switch t {
	case "h":
		return SplitHorizontal
	case "v":
		return SplitVertical
	}

	return "" // "p": a pane
}

func normalizeLayoutNode(n WindowLayoutNode, parent *WindowLayoutNode) LayoutNode {
	out := LayoutNode{Split: normalizeSplit(n.Type)}

	if parent != nil {
		out.Size = normalizeSize(n, *parent)
	}

	for _, c := range n.Children {
		out.Panes = append(out.Panes, normalizeLayoutNode(c, &n))
	}

	return out
}

func normalizeSize(n, parent WindowLayoutNode) int {
	if parent.Type == "v" {
		return percent(n.Height, parent.Height)
	}

	return percent(n.Width, parent.Width)
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

func percent(part, whole int) int {
	if whole <= 0 {
		return 0
	}

	return (part*100 + whole/2) / whole
}

func buildSetLayoutCmd(window Window) string {
	b, err := json.Marshal(window.WindowLayout) // []byte
	if err != nil {
		// NOTE: not user facing issue. This should never happen
		panic("invalid window layout")
	}

	target := utils.Quote(window.TargetIndex())
	return strings.Join([]string{"select-layout", "-t", target, utils.Quote(string(b))}, " ")
}
