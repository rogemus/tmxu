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

func (c Config) ShowTemplate(templateName string) ([]byte, error) {
	path, err := tmxuTemplatesDir(c.DirPath)
	if err != nil {
		return nil, err
	}

	templateFile := templateFile(path, templateName)
	out, err := readTemplateFile(templateFile)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c Config) SaveTemplate(data []byte, templateName string) (string, error) {
	path, err := tmxuTemplatesDir(c.DirPath)
	if err != nil {
		return "", err
	}

	templateFile := templateFile(path, templateName)
	return saveToFile(templateFile, data)
}

func (c Config) DeleteTemplate(templateName string) (string, error) {
	path, err := tmxuTemplatesDir(c.DirPath)
	if err != nil {
		return "", err
	}

	templateFile := templateFile(path, templateName)
	return deleteFile(templateFile)
}

func templateFile(path, name string) string {
	return filepath.Join(path, name+".json")
}

func readTemplatesName(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))

	for _, e := range entries {
		templateName := strings.TrimSuffix(e.Name(), ".json")
		names = append(names, templateName)
	}

	return names
}

func readTemplateFile(path string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadTemplateFile, path, err)
	}

	return file, nil
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
	configPath, err := tmxuConfigDir(dirPath)
	if err != nil {
		return "", err
	}

	if hasTemplatesDir(configPath) {
		path := tmxuTemplatesDirPath(configPath)
		return path, nil
	}

	path, err := createTmxuTemplatesDir(configPath)
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
