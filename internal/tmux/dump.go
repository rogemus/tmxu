package tmux

import (
	"encoding/json"
	"fmt"
	"strings"
)

type dumpData struct {
	Sessions []dumpSession `json:"sessions"`
}

type dumpSession struct {
	Name    string       `json:"name"`
	Windows []dumpWindow `json:"windows"`
}

type dumpWindow struct {
	Name   string      `json:"name"`
	Index  int         `json:"index"`
	Layout *LayoutNode `json:"layout,omitempty"` // split windows only
	Panes  []dumpPane  `json:"panes"`
}

type dumpPane struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

func (d Data) JSON() ([]byte, error) {
	out := dumpData{Sessions: make([]dumpSession, 0, len(d.sessions))}

	for _, sid := range d.sessionIDs() {
		s := dumpSession{
			Name:    d.sessions[sid].Name,
			Windows: make([]dumpWindow, 0, len(d.windows[sid])),
		}

		for _, w := range d.windows[sid] {
			dw := dumpWindow{
				Name:  w.Name,
				Index: w.Index,
				Panes: make([]dumpPane, 0, len(d.panes[w.ID])),
			}

			if len(w.WindowLayout) > 0 {
				layout, err := w.Layout()
				if err != nil {
					return nil, fmt.Errorf("window %s: %w", w.ID, err)
				}

				dw.Layout = &layout
			}

			for _, p := range d.panes[w.ID] {
				dw.Panes = append(dw.Panes, dumpPane{Title: p.Title, Path: p.Path})
			}

			s.Windows = append(s.Windows, dw)
		}

		out.Sessions = append(out.Sessions, s)
	}

	return json.MarshalIndent(out, "", "  ")
}

func (d Data) String() string {
	ids := d.sessionIDs()

	var b strings.Builder
	fmt.Fprintf(&b, "TmuxData: %d sessions\n", len(ids))

	for _, sid := range ids {
		fmt.Fprintf(&b, "%s %q\n", sid, d.sessions[sid].Name)

		for _, w := range d.windows[sid] {
			fmt.Fprintf(&b, "  %s [%d] %q\n", w.ID, w.Index, w.Name)

			if w.WindowLayout != nil {
				b.WriteString("    layout\n")

				layout, err := w.Layout()
				if err != nil {
					fmt.Fprintf(&b, "      <%v>\n", err)
				} else {
					writeLayoutNode(&b, layout, 3)
				}
			}

			for _, p := range d.panes[w.ID] {
				fmt.Fprintf(&b, "    %s [%d] %q %s\n", p.ID, p.Index, p.Title, p.Path)
			}
		}
	}

	return b.String()
}

func writeLayoutNode(b *strings.Builder, n LayoutNode, depth int) {
	kind := "pane"
	if n.Split != "" {
		kind = n.Split // "horizontal" or "vertical"
	}

	fmt.Fprintf(b, "%s%s %d%%\n", strings.Repeat("  ", depth), kind, n.Size)

	for _, p := range n.Panes {
		writeLayoutNode(b, p, depth+1)
	}
}
