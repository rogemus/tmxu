package cli

import "testing"

func TestPaneName(t *testing.T) {
	tests := []struct {
		name       string
		windowName string
		order      int16
		paneCount  int
		want       string
	}{
		{"single pane uses window name", "ai", 1, 1, "ai"},
		{"two panes first", "dev", 1, 2, "dev_1"},
		{"two panes second", "dev", 2, 2, "dev_2"},
		{"three panes third", "dev", 3, 3, "dev_3"},
		{"colon in window name kept", "dev:scripts", 1, 1, "dev:scripts"},
		{"colon in window name multi pane", "dev:scripts", 2, 2, "dev:scripts_2"},
		{"zero count treated as single", "ai", 1, 0, "ai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paneName(tt.windowName, tt.order, tt.paneCount)
			if got != tt.want {
				t.Errorf("paneName(%q, %d, %d) = %q, want %q", tt.windowName, tt.order, tt.paneCount, got, tt.want)
			}
		})
	}
}

func TestNewTPane(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		paneCount int
		wantName  string
		wantOrder int16
		wantPath  string
		wantErr   bool
	}{
		{"claude title ignored", "1\t/p\t✳ Plan mode off", 1, "ai", 1, "/p", false},
		{"hostname title ignored", "1\t/p\tDT-GR6CL4WVPC", 1, "ai", 1, "/p", false},
		{"empty title", "1\t/p\t", 1, "ai", 1, "/p", false},
		{"multi pane suffixed", "2\t/q\tzsh", 2, "ai_2", 2, "/q", false},
		{"bad order", "x\t/p\ttitle", 1, "", 0, "", true},
		{"too few fields", "1\t/p", 1, "", 0, "", true},
		{"empty input", "", 1, "", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTPane(tt.input, "sess", "sess:1", "ai", tt.paneCount)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Order != tt.wantOrder {
				t.Errorf("Order = %d, want %d", got.Order, tt.wantOrder)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", got.Path, tt.wantPath)
			}
			if got.SessionName != "sess" || got.SessionWindow != "sess:1" {
				t.Errorf("session fields = %q/%q", got.SessionName, got.SessionWindow)
			}
		})
	}
}
