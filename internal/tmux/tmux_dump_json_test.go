package tmux

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDataJSON(t *testing.T) {
	out, err := parseTmuxLines(decodeLines(t, layoutOutput)).JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	if !json.Valid(out) {
		t.Fatalf("JSON() is not valid JSON:\n%s", out)
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	sessions := got["sessions"].([]any)
	if len(sessions) != 1 {
		t.Fatalf("%d sessions, want 1", len(sessions))
	}

	session := sessions[0].(map[string]any)
	if session["name"] != "ran_liftup" {
		t.Errorf("session name = %v, want ran_liftup", session["name"])
	}

	windows := session["windows"].([]any)
	if len(windows) != 3 {
		t.Fatalf("%d windows, want 3", len(windows))
	}

	t.Run("split window has a normalized layout", func(t *testing.T) {
		w := windows[0].(map[string]any)

		wantLayout := map[string]any{
			"split": "horizontal",
			"size":  float64(100),
			"panes": []any{
				map[string]any{"size": float64(50)},
				map[string]any{"size": float64(50)},
			},
		}
		if !reflect.DeepEqual(w["layout"], wantLayout) {
			t.Errorf("layout = %v, want %v", w["layout"], wantLayout)
		}
		if w["name"] != "dev:scripts" || w["index"] != float64(2) {
			t.Errorf("window = %v, want dev:scripts at index 2", w)
		}

		panes := w["panes"].([]any)
		if len(panes) != 2 {
			t.Fatalf("%d panes, want 2", len(panes))
		}
		if first := panes[0].(map[string]any); first["title"] != "dev:scripts_1" || first["path"] != "/Users/test/liftup" {
			t.Errorf("first pane = %v", first)
		}
	})

	t.Run("single-pane windows have no layout", func(t *testing.T) {
		for _, i := range []int{1, 2} {
			w := windows[i].(map[string]any)

			if _, ok := w["layout"]; ok {
				t.Errorf("window %v has a layout, want none", w["name"])
			}
			if len(w["panes"].([]any)) != 1 {
				t.Errorf("window %v has %d panes, want 1", w["name"], len(w["panes"].([]any)))
			}
		}
	})

	t.Run("no runtime ids or tmux key names", func(t *testing.T) {
		for _, banned := range []string{`"%34"`, `"@34"`, `"$8"`, `"window_layout"`, `"V"`, `"L"`} {
			if strings.Contains(string(out), banned) {
				t.Errorf("JSON contains %s:\n%s", banned, out)
			}
		}
	})
}

func TestDataJSONEmpty(t *testing.T) {
	out, err := parseTmuxLines(nil).JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	var got struct {
		Sessions []any `json:"sessions"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.Sessions == nil || len(got.Sessions) != 0 {
		t.Errorf("sessions = %v, want an empty array (not null)\n%s", got.Sessions, out)
	}
}

func TestDataJSONInvalidLayout(t *testing.T) {
	d := Data{
		sessions: map[string]Session{"s1": {ID: "s1", Name: "a"}},
		windows: map[string][]Window{"s1": {
			{ID: "w1", Name: "w", WindowLayout: json.RawMessage(`d2b3,198x53,0,0,1`)},
		}},
		panes: map[string][]Pane{"w1": {{ID: "p1"}}},
	}

	out, err := d.JSON()

	if !errors.Is(err, ErrCannotParseLayout) {
		t.Fatalf("JSON() error = %v, want %v", err, ErrCannotParseLayout)
	}
	if !strings.Contains(err.Error(), "w1") {
		t.Errorf("error %q does not name the window", err)
	}
	if out != nil {
		t.Errorf("JSON() = %s, want no output on error", out)
	}
}
