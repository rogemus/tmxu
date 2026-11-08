// Package config holds the app configuration, such as the tmxu version and the
// path where saved sessions live, so it can be passed to commands.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrCannotReadConfigDir = errors.New("cannot read config dir")
	ErrCannotReadHomeDir   = errors.New("cannot read home dir")
)

const configDirName = "tmxu"

type Config struct {
	DirPath string
	Version string
}

func New(version string) (Config, error) {
	v := strings.TrimSpace(version)

	cDir, err := configDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		DirPath: cDir,
		Version: v,
	}, nil
}

func configDir() (string, error) {
	if runtime.GOOS == "darwin" {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg, nil
		}

		hDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrCannotReadHomeDir, err)
		}

		return filepath.Join(hDir, ".config"), nil
	}

	cDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotReadConfigDir, err)
	}

	return cDir, nil
}
