package tmux

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

func (d SnapshotData) JSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d SnapshotData) String() string {
	var b strings.Builder

	sessions := slices.Clone(d.Sessions)
	slices.SortFunc(sessions, func(x, y SessionData) int { return cmp.Compare(x.Name, y.Name) })

	for i, s := range sessions {
		if i > 0 {
			b.WriteString("\n")
		}

		fmt.Fprintln(&b, s.Name)
		writeWindows(&b, s)
	}

	return strings.TrimRight(b.String(), "\n")
}

func writeWindows(b *strings.Builder, s SessionData) {
	windows := slices.Clone(s.Windows)
	slices.SortFunc(windows, func(x, y WindowData) int { return cmp.Compare(x.Index, y.Index) })

	nameWidth := 0
	for _, w := range windows {
		for _, p := range w.Panes {
			nameWidth = max(nameWidth, utf8.RuneCountInString(paneTitle(w.Name, p.Index)))
		}
	}

	for wi, w := range windows {
		branch, rail := "├─ ", "│  "
		if wi == len(windows)-1 {
			branch, rail = "└─ ", "   "
		}

		fmt.Fprintf(b, "%s%s\n", branch, w.Name)

		for pi, p := range w.Panes {
			paneBranch := "├─ "
			if pi == len(w.Panes)-1 {
				paneBranch = "└─ "
			}

			writePane(b, rail+paneBranch, paneTitle(w.Name, p.Index), p.Path, nameWidth)
		}
	}
}

func writePane(b *strings.Builder, prefix, title, path string, nameWidth int) {
	if path == "" {
		fmt.Fprintf(b, "%s%s\n", prefix, title)
		return
	}

	fmt.Fprintf(b, "%s%-*s  %s\n", prefix, nameWidth, title, path)
}
