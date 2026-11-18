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

// String renders the snapshot as a tree of sessions, windows and panes.
// Panes show their name and start directory; ids, indexes and layouts are
// only part of the JSON output.
func (d SnapshotData) String() string {
	var b strings.Builder

	for i, sid := range d.sessionIDs() {
		if i > 0 {
			b.WriteString("\n")
		}

		fmt.Fprintln(&b, d.Sessions[sid].Name)
		d.writeWindows(&b, sid)
	}

	return strings.TrimRight(b.String(), "\n")
}

func (d SnapshotData) writeWindows(b *strings.Builder, sid string) {
	windows := slices.Clone(d.Windows[sid])
	slices.SortFunc(windows, func(x, y Window) int { return cmp.Compare(x.Index, y.Index) })

	// Window.Slug is not persisted, so a loaded snapshot has it empty;
	// derive the panes key the same way the snapshot keys were built.
	sessionName := d.Sessions[sid].Name
	panesOf := func(w Window) []Pane { return d.Panes[windowSlug(w, sessionName)] }

	nameWidth := 0
	for _, w := range windows {
		for _, p := range panesOf(w) {
			nameWidth = max(nameWidth, utf8.RuneCountInString(p.Title))
		}
	}

	for wi, w := range windows {
		branch, rail := "├─ ", "│  "
		if wi == len(windows)-1 {
			branch, rail = "└─ ", "   "
		}

		fmt.Fprintf(b, "%s%s\n", branch, w.Name)

		panes := panesOf(w)
		for pi, p := range panes {
			paneBranch := "├─ "
			if pi == len(panes)-1 {
				paneBranch = "└─ "
			}

			writePane(b, rail+paneBranch, p, nameWidth)
		}
	}
}

func writePane(b *strings.Builder, prefix string, p Pane, nameWidth int) {
	if p.Path == "" {
		fmt.Fprintf(b, "%s%s\n", prefix, p.Title)
		return
	}

	fmt.Fprintf(b, "%s%-*s  %s\n", prefix, nameWidth, p.Title, p.Path)
}
