package config

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewTrimsVersion(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "plain", version: "1.2.3", want: "1.2.3"},
		{name: "trailing newline from the embedded file", version: "1.2.3\n", want: "1.2.3"},
		{name: "surrounding spaces", version: "  1.2.3 ", want: "1.2.3"},
		{name: "empty", version: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New(tt.version)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if cfg.Version != tt.want {
				t.Errorf("Version = %q, want %q", cfg.Version, tt.want)
			}
		})
	}
}

func TestNewDirPath(t *testing.T) {
	t.Run("XDG_CONFIG_HOME wins", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv("HOME", t.TempDir())

		cfg, err := New("1.0.0")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if want := filepath.Join(xdg, "tmxu"); cfg.DirPath != want {
			t.Errorf("DirPath = %q, want %q", cfg.DirPath, want)
		}
	})

	t.Run("falls back to ~/.config", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", home)

		cfg, err := New("1.0.0")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if want := filepath.Join(home, ".config", "tmxu"); cfg.DirPath != want {
			t.Errorf("DirPath = %q, want %q", cfg.DirPath, want)
		}
	})
}

func TestNewWithoutHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")

	cfg, err := New("1.0.0")
	if err == nil {
		t.Fatalf("New() = %+v, want an error", cfg)
	}

	// macOS resolves ~/.config itself, other systems ask os.UserConfigDir.
	want := ErrCannotReadConfigDir
	if runtime.GOOS == "darwin" {
		want = ErrCannotReadHomeDir
	}

	if !errors.Is(err, want) {
		t.Errorf("error = %v, want it to wrap %v", err, want)
	}
	if cfg != (Config{}) {
		t.Errorf("Config = %+v, want zero value on error", cfg)
	}
}
