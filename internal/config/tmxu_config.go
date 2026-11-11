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
	ErrCannotReadFile            = errors.New("cannot read file")
	ErrCannotDeleteFile          = errors.New("cannot delete file")
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

func readFile(root *os.Root, name string) ([]byte, error) {
	file, err := root.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadFile, name, err)
	}

	return file, nil
}

func saveToFile(root *os.Root, name string, data []byte) (string, error) {
	err := root.WriteFile(name, data, 0644)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, name, err)
	}

	return filepath.Join(root.Name(), name), nil
}

func deleteFile(root *os.Root, name string) (string, error) {
	err := root.Remove(name)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotDeleteFile, name, err)
	}

	return filepath.Join(root.Name(), name), nil
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
