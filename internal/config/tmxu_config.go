package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrCannotSaveFile            = errors.New("cannot save file")
	ErrCannotCreateTmxuConfigDir = errors.New("cannot create tmxu config dir")
	ErrCannotReadFile            = errors.New("cannot read file")
	ErrCannotDeleteFile          = errors.New("cannot delete file")
	ErrCannotCheckFile           = errors.New("cannot check file")
)

func tmxuConfigDir(dirPath string) (string, error) {
	path := tmxuConfigDirPath(dirPath)
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotCreateTmxuConfigDir, err)
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

func hasFile(root *os.Root, name string) (bool, error) {
	_, err := root.Stat(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("%w %s: %w", ErrCannotCheckFile, name, err)
	}

	return true, nil
}

func saveToFile(root *os.Root, name string, data []byte) (string, error) {
	err := root.WriteFile(name, data, 0600)
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
