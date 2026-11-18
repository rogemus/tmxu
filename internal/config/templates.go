package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrInvalidTemplateName          = errors.New("invalid template name")
	ErrCannotReadTemplatesDir       = errors.New("cannot read templates dir")
	ErrCannotCreateTmxuTemplatesDir = errors.New("cannot create tmxu templates dir")
)

const templatesDirName = "templates"

func (c Config) TemplatesDir() string {
	return tmxuTemplatesDirPath(c.AppDirPath)
}

func (c Config) ListTemplates() ([]string, error) {
	path := tmxuTemplatesDirPath(c.AppDirPath)

	entries, err := readTemplateFiles(path)
	if err != nil {
		return nil, err
	}

	return readTemplatesName(entries), nil
}

func (c Config) ShowTemplate(templateName string) ([]byte, error) {
	fileName, err := templateFileName(templateName)
	if err != nil {
		return nil, err
	}

	path := tmxuTemplatesDirPath(c.AppDirPath)

	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	return readFile(root, fileName)
}

func (c Config) SaveTemplate(data []byte, templateName string) (string, error) {
	fileName, err := templateFileName(templateName)
	if err != nil {
		return "", err
	}

	path, err := tmxuTemplatesDir(c.AppDirPath)
	if err != nil {
		return "", err
	}

	root, err := os.OpenRoot(path)
	if err != nil {
		return "", err
	}
	defer root.Close()

	return saveToFile(root, fileName, data)
}

func (c Config) DeleteTemplate(templateName string) (string, error) {
	fileName, err := templateFileName(templateName)
	if err != nil {
		return "", err
	}

	path := tmxuTemplatesDirPath(c.AppDirPath)

	root, err := os.OpenRoot(path)
	if err != nil {
		return "", err
	}
	defer root.Close()

	return deleteFile(root, fileName)
}

func (c Config) HasTemplate(templateName string) (bool, error) {
	fileName, err := templateFileName(templateName)
	if err != nil {
		return false, err
	}

	path := tmxuTemplatesDirPath(c.AppDirPath)

	root, err := os.OpenRoot(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer root.Close()

	return hasFile(root, fileName)
}

func templateFileName(name string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(name))
	if trimmed == "" {
		return "", fmt.Errorf("%w: %q", ErrInvalidTemplateName, name)
	}

	return utils.Slug(trimmed) + ".json", nil
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
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
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

func tmxuTemplatesDirPath(path string) string {
	return filepath.Join(path, templatesDirName)
}

func tmxuTemplatesDir(dirPath string) (string, error) {
	path := tmxuTemplatesDirPath(dirPath)
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotCreateTmxuTemplatesDir, err)
	}

	return path, nil
}
