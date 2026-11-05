package tmux

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (d SnapshotData) JSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d SnapshotData) String() string {
	ids := d.sessionIDs()

	var b strings.Builder
	fmt.Fprintf(&b, "TmuxData: %d sessions\n", len(ids))

	for _, sid := range ids {
		fmt.Fprintf(&b, "%s %q\n", sid, d.Sessions[sid].Name)

		for _, w := range d.Windows[sid] {
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

			for _, p := range d.Panes[w.Slug] {
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
