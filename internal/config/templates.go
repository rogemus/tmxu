package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCannotReadTemplatesDir       = errors.New("cannot read templates dir")
	ErrCannotReadTemplateFile       = errors.New("cannot read template file")
	ErrCannotSaveTemplate           = errors.New("cannot save template")
	ErrCannotDeleteTemplate         = errors.New("cannot delete template")
	ErrCannotCreateTmxuTemplatesDir = errors.New("cannot create tmxu templates dir")
)

const templatesDirName = "templates"

func (c Config) ListTemplates() ([]string, error) {
	path, err := tmxuConfigDir(c.DirPath)
	if err != nil {
		return nil, err
	}

	path, err = tmxuTemplatesDir(path)
	if err != nil {
		return nil, err
	}

	entries, err := readTemplateFiles(path)
	if err != nil {
		return nil, err
	}

	return readTemplatesName(entries), nil
}

func (c Config) SaveTemplate() error {
	return nil
}

func (c Config) DeleteTemplate() error {
	return nil
}

func readTemplatesName(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))

	for _, e := range entries {
		templateName := strings.TrimSuffix(e.Name(), ".json")
		names = append(names, templateName)
	}

	return names
}

func readTemplateFiles(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadTemplatesDir, path, err)
	}

	filtered := make([]os.DirEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		filtered = append(filtered, e)
	}

	return filtered, nil
}

func hasTemplatesDir(dirPath string) bool {
	path := tmxuTemplatesDirPath(dirPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func tmxuTemplatesDirPath(path string) string {
	return filepath.Join(path, templatesDirName)
}

func tmxuTemplatesDir(dirPath string) (string, error) {
	if hasTemplatesDir(dirPath) {
		path := tmxuTemplatesDirPath(dirPath)
		return path, nil
	}

	path, err := createTmxuTemplatesDir(dirPath)
	if err != nil {
		return "", err
	}

	return path, nil
}

func createTmxuTemplatesDir(dirPath string) (string, error) {
	path := tmxuTemplatesDirPath(dirPath)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotCreateTmxuTemplatesDir, err)
	}

	return path, nil

}
