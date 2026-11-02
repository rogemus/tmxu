package tmux

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/rogemus/tmxu/internal/utils"
)

func TestLineConversions(t *testing.T) {
	l := Line{
		SessionID:   "$1",
		SessionName: "test-session",
		WindowID:    "@2",
		WindowIndex: 3,
		WindowName:  "test-window",
		PaneID:      "%4",
		PaneIndex:   5,
		PanePath:    "/test/path",
		PaneTitle:   "test-title",
	}

	t.Run("Session", func(t *testing.T) {
		want := Session{ID: "$1", Name: "test-session"}
		if got := l.Session(); got != want {
			t.Errorf("Session() = %+v, want %+v", got, want)
		}
	})

	t.Run("Window carries its session name", func(t *testing.T) {
		want := Window{ID: "@2", Index: 3, Name: "test-window", SessionName: "test-session"}
		if got := l.Window(); !reflect.DeepEqual(got, want) {
			t.Errorf("Window() = %+v, want %+v", got, want)
		}
	})

	t.Run("Window keeps a split layout", func(t *testing.T) {
		split := l
		split.WindowLayout = json.RawMessage(`{"V":2,"L":{"t":"h","c":[{"t":"p"},{"t":"p"}]}}`)

		if got := split.Window().WindowLayout; string(got) != string(split.WindowLayout) {
			t.Errorf("Window().WindowLayout = %s, want %s", got, split.WindowLayout)
		}
	})

	t.Run("Window drops a single-pane layout", func(t *testing.T) {
		single := l
		single.WindowLayout = json.RawMessage(`{"V":2,"L":{"t":"p","w":10,"h":10}}`)

		if got := single.Window().WindowLayout; got != nil {
			t.Errorf("Window().WindowLayout = %s, want none", got)
		}
	})

	t.Run("Pane", func(t *testing.T) {
		want := Pane{ID: "%4", Index: 5, Title: "test-title", Path: "/test/path"}
		if got := l.Pane(); got != want {
			t.Errorf("Pane() = %+v, want %+v", got, want)
		}
	})
}

func TestWindowQuotingAndEq(t *testing.T) {
	t.Run("GetName quotes the name", func(t *testing.T) {
		tests := map[string]string{
			"plain":      "'plain'",
			"with space": "'with space'",
			"it's":       `'it'\''s'`,
			"":           "''",
		}
		for name, want := range tests {
			if got := (Window{Name: name}).GetName(); got != want {
				t.Errorf("GetName(%q) = %s, want %s", name, got, want)
			}
		}
	})

	w := Window{ID: "@1", Name: "a"}
	tests := []struct {
		name  string
		other utils.Comperable
		want  bool
	}{
		{name: "same id", other: Window{ID: "@1", Name: "other"}, want: true},
		{name: "different id", other: Window{ID: "@2", Name: "a"}, want: false},
		{name: "different type with same id", other: Pane{ID: "@1"}, want: false},
	}
	for _, tt := range tests {
		t.Run("Eq "+tt.name, func(t *testing.T) {
			if got := w.Eq(tt.other); got != tt.want {
				t.Errorf("Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTmuxLinesEmpty(t *testing.T) {
	for _, lines := range [][]Line{nil, {}} {
		got := parseTmuxLines(lines)

		if len(got.sessions) != 0 || len(got.windows) != 0 || len(got.panes) != 0 {
			t.Errorf("parseTmuxLines(%v) = %+v, want empty Data", lines, got)
		}
	}
}

func TestDataString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseTmuxLines(nil).String()
		if want := "TmuxData: 0 sessions\n"; got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	})

	t.Run("sessions, windows and panes are nested", func(t *testing.T) {
		d := parseTmuxLines([]Line{
			line("s1", "w1", "p1", 1, 1),
			line("s1", "w1", "p2", 1, 2),
		})

		want := "TmuxData: 1 sessions\n" +
			"s1 \"s1-name\"\n" +
			"  w1 [1] \"w1-name\"\n" +
			"    p1 [1] \"p1-title\" /path/p1\n" +
			"    p2 [2] \"p2-title\" /path/p2\n"

		if got := d.String(); got != want {
			t.Errorf("String() =\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("names with quotes and newlines are escaped by %q", func(t *testing.T) {
		l := line("s1", "w1", "p1", 1, 1)
		l.PaneTitle = "a \"b\"\nc"
		got := parseTmuxLines([]Line{l}).String()

		want := "    p1 [1] " + `"a \"b\"\nc"` + " /path/p1\n"
		if !strings.HasSuffix(got, want) {
			t.Errorf("String() = %q, want suffix %q", got, want)
		}
	})
}
