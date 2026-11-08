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

func tmxuConfigDir(dirPath string) (string, error) {
	if hasTmxuConfigDir(dirPath) {
		path := tmxuConfigDirPath(dirPath)
		return path, nil
	}

	path, err := createTmxuConfigDir(dirPath)
	if err != nil {
		return "", err
	}

	return path, nil
}

func tmxuConfigDirPath(path string) string {
	return filepath.Join(path, configDirName)
}

func saveToFile(path string, data []byte) (string, error) {
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, path, err)
	}

	return path, nil
}

func hasTmxuConfigDir(path string) bool {
	p := tmxuConfigDirPath(path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}

	return true
}

func createTmxuConfigDir(path string) (string, error) {
	p := tmxuConfigDirPath(path)
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotCreateTmxuConfigDir, err)
	}

	return p, nil
}
