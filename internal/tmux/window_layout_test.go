package tmux

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func layoutPane(w, h int) WindowLayoutNode {
	return WindowLayoutNode{Type: "p", Width: w, Height: h}
}

func layoutSplit(t string, w, h int, children ...WindowLayoutNode) WindowLayoutNode {
	return WindowLayoutNode{Type: t, Width: w, Height: h, Children: children}
}

func TestPercent(t *testing.T) {
	tests := []struct {
		name        string
		part, whole int
		want        int
	}{
		{name: "half", part: 50, whole: 100, want: 50},
		{name: "full", part: 100, whole: 100, want: 100},
		{name: "first of two panes", part: 190, whole: 380, want: 50},
		{name: "second of two panes rounds to the nearest", part: 189, whole: 380, want: 50},
		{name: "rounds down", part: 1, whole: 3, want: 33},
		{name: "rounds up", part: 2, whole: 3, want: 67},
		{name: "tall pane", part: 40, whole: 67, want: 60},
		{name: "short pane", part: 26, whole: 67, want: 39},
		{name: "zero part", part: 0, whole: 10, want: 0},
		{name: "zero whole", part: 5, whole: 0, want: 0},
		{name: "negative whole", part: 5, whole: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.part, tt.whole); got != tt.want {
				t.Errorf("percent(%d, %d) = %d, want %d", tt.part, tt.whole, got, tt.want)
			}
		})
	}
}

func TestNormalizeSplit(t *testing.T) {
	tests := map[string]string{
		"h":  SplitHorizontal,
		"v":  SplitVertical,
		"p":  "",
		"":   "",
		"x":  "",
		"hh": "",
	}

	for in, want := range tests {
		if got := normalizeSplit(in); got != want {
			t.Errorf("normalizeSplit(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeSize(t *testing.T) {
	tests := []struct {
		name   string
		node   WindowLayoutNode
		parent WindowLayoutNode
		want   int
	}{
		{
			name:   "side by side uses the width",
			node:   layoutPane(190, 67),
			parent: layoutSplit("h", 380, 67),
			want:   50,
		},
		{
			name:   "stacked uses the height",
			node:   layoutPane(198, 40),
			parent: layoutSplit("v", 198, 67),
			want:   60,
		},
		{
			name:   "unknown parent type falls back to the width",
			node:   layoutPane(100, 10),
			parent: layoutSplit("", 200, 20),
			want:   50,
		},
		{
			name:   "zero width parent does not divide by zero",
			node:   layoutPane(10, 10),
			parent: layoutSplit("h", 0, 10),
			want:   0,
		},
		{
			name:   "zero height parent does not divide by zero",
			node:   layoutPane(10, 10),
			parent: layoutSplit("v", 10, 0),
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSize(tt.node, tt.parent); got != tt.want {
				t.Errorf("normalizeSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeWindowLayout(t *testing.T) {
	tests := []struct {
		name   string
		layout WindowLayout
		want   LayoutNode
	}{
		{
			name:   "single pane is the whole window",
			layout: WindowLayout{Version: 2, Root: layoutPane(198, 53)},
			want:   LayoutNode{Size: 100},
		},
		{
			name: "two panes side by side",
			layout: WindowLayout{Root: layoutSplit("h", 380, 67,
				layoutPane(190, 67), layoutPane(189, 67))},
			want: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{
				{Size: 50}, {Size: 50},
			}},
		},
		{
			name: "two panes stacked use the height",
			layout: WindowLayout{Root: layoutSplit("v", 198, 67,
				layoutPane(198, 40), layoutPane(198, 26))},
			want: LayoutNode{Split: SplitVertical, Size: 100, Panes: []LayoutNode{
				{Size: 60}, {Size: 39},
			}},
		},
		{
			name: "three panes keep their order",
			layout: WindowLayout{Root: layoutSplit("h", 300, 50,
				layoutPane(99, 50), layoutPane(100, 50), layoutPane(99, 50))},
			want: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{
				{Size: 33}, {Size: 33}, {Size: 33},
			}},
		},
		{
			name: "nested split is relative to its own parent",
			layout: WindowLayout{Root: layoutSplit("h", 380, 67,
				layoutPane(190, 67),
				layoutSplit("v", 189, 67, layoutPane(189, 40), layoutPane(189, 26)),
			)},
			want: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{
				{Size: 50},
				{Split: SplitVertical, Size: 50, Panes: []LayoutNode{{Size: 60}, {Size: 39}}},
			}},
		},
		{
			name: "zero sizes do not panic",
			layout: WindowLayout{Root: layoutSplit("h", 0, 0,
				layoutPane(0, 0), layoutPane(0, 0))},
			want: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{{}, {}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWindowLayout(tt.layout)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeWindowLayout() =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}

func TestWindowLayoutUnmarshal(t *testing.T) {
	// Real tmux 3.8 text, including the "l" key the structs do not model.
	raw := `{"V":2,"L":{"t":"h","w":200,"h":50,"x":0,"y":0,"c":[` +
		`{"t":"p","w":100,"h":50,"x":0,"y":0,"l":0,"i":1,"I":"%0"},` +
		`{"t":"p","w":99,"h":50,"x":101,"y":0,"a":true,"i":2,"I":"%1"}]}}`

	var got WindowLayout
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	root := got.Root
	if got.Version != 2 || root.Type != "h" || root.Width != 200 || root.Height != 50 {
		t.Errorf("root = %+v (V=%d), want a 200x50 side-by-side split, V=2", root, got.Version)
	}
	if len(root.Children) != 2 {
		t.Fatalf("root has %d children, want 2", len(root.Children))
	}

	first, second := root.Children[0], root.Children[1]
	if first.Type != "p" || first.PaneID != "%0" || first.Index != 1 || first.Active || len(first.Children) != 0 {
		t.Errorf("first = %+v, want inactive leaf pane %%0 with index 1", first)
	}
	if second.PaneID != "%1" || second.Index != 2 || !second.Active || second.X != 101 {
		t.Errorf("second = %+v, want active pane %%1 with index 2 at x=101", second)
	}
}

func TestLayoutNodeMarshal(t *testing.T) {
	tests := []struct {
		name string
		node LayoutNode
		want string
	}{
		{
			name: "pane has only a size",
			node: LayoutNode{Size: 50},
			want: `{"size":50}`,
		},
		{
			name: "split lists its panes",
			node: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{{Size: 50}, {Size: 50}}},
			want: `{"split":"horizontal","size":100,"panes":[{"size":50},{"size":50}]}`,
		},
		{
			name: "nested split",
			node: LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{
				{Size: 50},
				{Split: SplitVertical, Size: 50, Panes: []LayoutNode{{Size: 60}, {Size: 40}}},
			}},
			want: `{"split":"horizontal","size":100,"panes":[{"size":50},` +
				`{"split":"vertical","size":50,"panes":[{"size":60},{"size":40}]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.node)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestIsSplitLayout(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want bool
	}{
		{name: "split", raw: json.RawMessage(`{"V":2,"L":{"t":"h","c":[{"t":"p"},{"t":"p"}]}}`), want: true},
		{name: "single pane", raw: json.RawMessage(`{"V":2,"L":{"t":"p","w":10,"h":10}}`), want: false},
		{name: "nil", raw: nil, want: false},
		{name: "empty", raw: json.RawMessage(``), want: false},
		{name: "null", raw: json.RawMessage(`null`), want: false},
		{name: "not json", raw: json.RawMessage(`d2b3,198x53,0,0,1`), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSplitLayout(tt.raw); got != tt.want {
				t.Errorf("isSplitLayout(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestWindowLayoutMethod(t *testing.T) {
	t.Run("valid layout is normalized", func(t *testing.T) {
		w := Window{WindowLayout: json.RawMessage(
			`{"V":2,"L":{"t":"h","w":380,"h":67,"c":[{"t":"p","w":190,"h":67},{"t":"p","w":189,"h":67}]}}`)}

		got, err := w.Layout()
		if err != nil {
			t.Fatalf("Layout() error = %v", err)
		}

		want := LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{{Size: 50}, {Size: 50}}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Layout() = %+v, want %+v", got, want)
		}
	})

	for name, raw := range map[string]json.RawMessage{
		"empty":    nil,
		"not json": json.RawMessage(`d2b3,198x53,0,0,1`),
	} {
		t.Run(name+" layout is an error", func(t *testing.T) {
			_, err := Window{WindowLayout: raw}.Layout()

			if !errors.Is(err, ErrCannotParseLayout) {
				t.Errorf("Layout() error = %v, want %v", err, ErrCannotParseLayout)
			}
		})
	}
}

func TestWriteLayoutNode(t *testing.T) {
	node := LayoutNode{Split: SplitHorizontal, Size: 100, Panes: []LayoutNode{
		{Size: 50},
		{Split: SplitVertical, Size: 50, Panes: []LayoutNode{{Size: 60}, {Size: 39}}},
	}}

	tests := []struct {
		name  string
		depth int
		want  string
	}{
		{
			name:  "depth 0",
			depth: 0,
			want: "horizontal 100%\n" +
				"  pane 50%\n" +
				"  vertical 50%\n" +
				"    pane 60%\n" +
				"    pane 39%\n",
		},
		{
			name:  "depth 3 adds six spaces",
			depth: 3,
			want: "      horizontal 100%\n" +
				"        pane 50%\n" +
				"        vertical 50%\n" +
				"          pane 60%\n" +
				"          pane 39%\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder

			writeLayoutNode(&b, node, tt.depth)

			if b.String() != tt.want {
				t.Errorf("writeLayoutNode() =\n%s\nwant\n%s", b.String(), tt.want)
			}
		})
	}
}

func TestBuildSetLayoutCmd(t *testing.T) {
	tests := []struct {
		name   string
		window Window
		want   string
	}{
		{
			name: "layout is quoted and targets the exact window",
			window: Window{
				Index: 2, SessionName: "s",
				WindowLayout: json.RawMessage(`{"V":2,"L":{"t":"p"}}`),
			},
			want: `select-layout -t '=s:2' '{"V":2,"L":{"t":"p"}}'`,
		},
		{
			name: "session name with a space",
			window: Window{
				Index: 1, SessionName: "my session",
				WindowLayout: json.RawMessage(`{"V":2}`),
			},
			want: `select-layout -t '=my session:1' '{"V":2}'`,
		},
		{
			name: "single quote inside the layout is escaped",
			window: Window{
				Index: 1, SessionName: "s",
				WindowLayout: json.RawMessage(`{"a":"it's"}`),
			},
			want: `select-layout -t '=s:1' '{"a":"it'\''s"}'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSetLayoutCmd(tt.window); got != tt.want {
				t.Errorf("buildSetLayoutCmd() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestBuildNewWindowsCmdAppliesLayoutAfterPanes(t *testing.T) {
	window := Window{
		ID: "w1", Index: 1, Name: "dev", SessionName: "s",
		WindowLayout: json.RawMessage(`{"V":2,"L":{"t":"h"}}`),
	}
	panes := map[string][]Pane{"w1": {{Path: "/a", Title: "t1"}, {Path: "/b", Title: "t2"}}}

	got := buildNewWindowsCmd([]Window{window}, panes)

	if !strings.HasSuffix(got, `select-layout -t '=s:1' '{"V":2,"L":{"t":"h"}}'`) {
		t.Errorf("layout is not the last command:\n%s", got)
	}
	if strings.Index(got, "split-window") > strings.Index(got, "select-layout") {
		t.Errorf("select-layout must come after the splits:\n%s", got)
	}

	window.WindowLayout = nil
	if got := buildNewWindowsCmd([]Window{window}, panes); strings.Contains(got, "select-layout") {
		t.Errorf("a window without a layout got select-layout:\n%s", got)
	}
}
