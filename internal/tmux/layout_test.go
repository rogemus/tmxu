package tmux

import (
	"encoding/json"
	"strings"
	"testing"
)

// layoutOutput is real `tmux list-panes -a -F` output (tmux 3.8) for one
// session with a two-pane window and two single-pane windows.
const layoutOutput = `{"session_id": "$8","session_name": "ran_liftup","window_id": "@34","window_index": 2,"window_name": "dev:scripts","window_layout": {"V":2,"L":{"t":"h","w":380,"h":67,"x":0,"y":0,"c":[{"t":"p","w":190,"h":67,"x":0,"y":0,"a":true,"i":1,"I":"%34"},{"t":"p","w":189,"h":67,"x":191,"y":0,"i":2,"I":"%35"}]}},"pane_id": "%34","pane_index": 1,"pane_path": "/Users/test/liftup","pane_title": "dev:scripts_1"},
{"session_id": "$8","session_name": "ran_liftup","window_id": "@34","window_index": 2,"window_name": "dev:scripts","window_layout": {"V":2,"L":{"t":"h","w":380,"h":67,"x":0,"y":0,"c":[{"t":"p","w":190,"h":67,"x":0,"y":0,"a":true,"i":1,"I":"%34"},{"t":"p","w":189,"h":67,"x":191,"y":0,"i":2,"I":"%35"}]}},"pane_id": "%35","pane_index": 2,"pane_path": "/Users/test/liftup","pane_title": "dev:scripts_2"},
{"session_id": "$8","session_name": "ran_liftup","window_id": "@35","window_index": 3,"window_name": "lzg","window_layout": {"V":2,"L":{"t":"p","w":380,"h":67,"x":0,"y":0,"a":true,"i":1,"I":"%36"}},"pane_id": "%36","pane_index": 1,"pane_path": "/Users/test/liftup","pane_title": "lzg"},
{"session_id": "$8","session_name": "ran_liftup","window_id": "@36","window_index": 4,"window_name": "ai","window_layout": {"V":2,"L":{"t":"p","w":198,"h":53,"x":0,"y":0,"a":true,"i":1,"I":"%37"}},"pane_id": "%37","pane_index": 1,"pane_path": "/Users/test/liftup","pane_title": "ai"},
`

// decodeLines mirrors how GetTmuxData turns tmux output into lines.
func decodeLines(t *testing.T, out string) []Line {
	t.Helper()

	doc := "[" + strings.TrimSuffix(strings.TrimSpace(out), ",") + "]"

	var lines []Line
	if err := json.Unmarshal([]byte(doc), &lines); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return lines
}

func TestLineDecodesLayout(t *testing.T) {
	lines := decodeLines(t, layoutOutput)

	if len(lines) != 4 {
		t.Fatalf("decoded %d lines, want 4", len(lines))
	}

	// The layout is kept as the exact text tmux printed.
	wantRaw := `{"V":2,"L":{"t":"h","w":380,"h":67,"x":0,"y":0,"c":[` +
		`{"t":"p","w":190,"h":67,"x":0,"y":0,"a":true,"i":1,"I":"%34"},` +
		`{"t":"p","w":189,"h":67,"x":191,"y":0,"i":2,"I":"%35"}]}}`
	if got := string(lines[0].WindowLayout); got != wantRaw {
		t.Errorf("WindowLayout = %s, want %s", got, wantRaw)
	}

	var split, single WindowLayout
	if err := json.Unmarshal(lines[0].WindowLayout, &split); err != nil {
		t.Fatalf("Unmarshal(split) error = %v", err)
	}
	if err := json.Unmarshal(lines[2].WindowLayout, &single); err != nil {
		t.Fatalf("Unmarshal(single) error = %v", err)
	}

	if len(split.Root.Children) != 2 {
		t.Errorf("split window has %d children, want 2", len(split.Root.Children))
	}
	if single.Root.Type != "p" || len(single.Root.Children) != 0 {
		t.Errorf("single-pane root = %+v, want a leaf", single.Root)
	}
}

func TestDataStringShowsLayoutOnlyForSplitWindows(t *testing.T) {
	got := parseTmuxLines(decodeLines(t, layoutOutput)).String()

	wantLayout := "    layout\n" +
		"      horizontal 100%\n" +
		"        pane 50%\n" +
		"        pane 50%\n"

	if n := strings.Count(got, "layout\n"); n != 1 {
		t.Errorf("String() has %d layout blocks, want 1 (only the split window):\n%s", n, got)
	}

	// The layout block sits between its window and that window's panes.
	want := "  @34 [2] \"dev:scripts\"\n" + wantLayout + "    %34 [1]"
	if !strings.Contains(got, want) {
		t.Errorf("String() does not contain\n%s\nin\n%s", want, got)
	}
}

func TestParseTmuxLinesKeepsLayoutOnlyForSplitWindows(t *testing.T) {
	data := parseTmuxLines(decodeLines(t, layoutOutput))

	windows := data.windows["$8"]
	if len(windows) != 3 {
		t.Fatalf("session has %d windows, want 3", len(windows))
	}

	tests := []struct {
		id         string
		wantLayout bool
		wantPanes  int
	}{
		{id: "@34", wantLayout: true, wantPanes: 2},
		{id: "@35", wantLayout: false, wantPanes: 1},
		{id: "@36", wantLayout: false, wantPanes: 1},
	}

	for i, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			w := windows[i]
			if w.ID != tt.id {
				t.Fatalf("windows[%d] = %s, want %s", i, w.ID, tt.id)
			}
			if got := w.WindowLayout != nil; got != tt.wantLayout {
				t.Errorf("layout present = %v, want %v", got, tt.wantLayout)
			}
			if got := len(data.panes[tt.id]); got != tt.wantPanes {
				t.Errorf("%d panes, want %d", got, tt.wantPanes)
			}
		})
	}

	layout, err := windows[0].Layout()
	if err != nil {
		t.Fatalf("Layout() error = %v", err)
	}
	if len(layout.Panes) != 2 {
		t.Errorf("split window layout has %d panes, want 2", len(layout.Panes))
	}
}
