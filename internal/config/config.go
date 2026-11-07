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
// TODO: temp name remove 2 once everything is working
const sessionsFileName = "tmux-sessions2.json"

type Config struct {
	DirPath string
	Version string
}

func New(version string) (Config, error) {
	v := strings.TrimSpace(version)

	dir, err := configDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		DirPath: dir,
		Version: v,
	}, nil
}

func (c Config) SessionsFileName() string {
	return fmt.Sprintf("%s/%s", c.DirPath, sessionsFileName)
}

func configDir() (string, error) {
	if runtime.GOOS == "darwin" {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, configDirName), nil
		}

		hDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrCannotReadHomeDir, err)
		}

		return filepath.Join(hDir, ".config", configDirName), nil
	}

	cDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotReadConfigDir, err)
	}
	return filepath.Join(cDir, configDirName), nil
}
