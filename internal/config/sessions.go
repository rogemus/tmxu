package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrCannotSaveFile            = errors.New("cannot save file")
	ErrCannotCreateTmxuConfigDir = errors.New("cannot create tmxu config dir")
)

func (c Config) SaveSessions(data []byte) (string, error) {
	var (
		err          error
		path         string
		sessionsFile string
	)

	if hasTmxuConfigDir(c.DirPath) {
		path = tmxuConfigDir(c.DirPath)
		sessionsFile = tmxuSessionFile(path)

		return saveToFile(sessionsFile, data)
	}

	path, err = createTmxuConfigDir(c.DirPath)
	if err != nil {
		return "", err
	}

	sessionsFile = tmxuSessionFile(path)
	return saveToFile(sessionsFile, data)
}

func tmxuConfigDir(path string) string {
	return filepath.Join(path, configDirName)
}

func tmxuSessionFile(path string) string {
	return filepath.Join(path, sessionsFileName)
}

func saveToFile(path string, data []byte) (string, error) {
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, path, err)
	}

	return path, nil
}

func hasTmxuConfigDir(path string) bool {
	p := tmxuConfigDir(path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}

	return true
}

func createTmxuConfigDir(path string) (string, error) {
	p := tmxuConfigDir(path)
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotCreateTmxuConfigDir, err)
	}

	return p, nil
}
