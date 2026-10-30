package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrCannotCheckConfigDir     = errors.New("cannot check config dir")
	ErrCannotCreateConfigDir    = errors.New("cannot create config dir")
	ErrCannotCreateTemplatesDir = errors.New("cannot create templates dir")
	ErrCannotGetHomeDir         = errors.New("cannot get home dir")
	ErrCannotGetFilePath        = errors.New("cannot get file path")
	ErrCannotGetTemplatesDir    = errors.New("cannot get templates dir")
	ErrCannotMarshalData        = errors.New("cannot marshal data")
	ErrCannotUnmarshalData      = errors.New("cannot unmarshal data")
	ErrCannotSaveSessionFile    = errors.New("cannot save session file")
	ErrCannotReadSessionFile    = errors.New("cannot read session file")
	ErrCannotReadTemplatesDir   = errors.New("cannot read templates dir")
	ErrCannotReadTemplateFile   = errors.New("cannot read template file")
	ErrCannotSaveTemplate       = errors.New("cannot save template")
	ErrCannotDeleteTemplate     = errors.New("cannot delete template")
)

const configDir = "/.config/tmxu/"
const templatesDir = "templates"
const sessionFile = "tmux-sessions.json"

func saveSessionsFile(data []tSession) error {
	hasConfigDir, err := hasConfigDir()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotCheckConfigDir, configDir, err)
	}

	if !hasConfigDir {
		if err := createConfigDir(); err != nil {
			return fmt.Errorf("%w %s: %w", ErrCannotCreateConfigDir, configDir, err)
		}
	}

	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotMarshalData, err)
	}

	path, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotGetFilePath, err)
	}

	err = os.WriteFile(path, j, 0644)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotSaveSessionFile, path, err)
	}

	return nil
}

func hasConfigDir() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCannotGetHomeDir, err)
	}

	path := filepath.Join(homeDir, configDir)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func createConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotGetHomeDir, err)
	}

	path := filepath.Join(homeDir, configDir)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotCreateConfigDir, path, err)
	}

	return nil
}

func loadSessionsFile() ([]tSession, error) {
	var data []tSession

	path, err := getSessionFilePath()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotGetFilePath, err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadSessionFile, path, err)
	}

	err = json.Unmarshal(out, &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotUnmarshalData, err)
	}

	return data, nil
}

func getSessionFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotGetHomeDir, err)
	}

	return filepath.Join(homeDir, configDir, sessionFile), nil
}

func hasTemplatesDir() (bool, error) {
	_, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCannotGetHomeDir, err)
	}

	path, _ := getTemplatesDirPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func createTemplatesDir() error {
	_, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotGetHomeDir, err)
	}

	path, _ := getTemplatesDirPath()
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotCreateTemplatesDir, path, err)
	}

	return nil
}

func getTemplatesDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotGetTemplatesDir, err)
	}

	return filepath.Join(homeDir, configDir, templatesDir), nil
}

func loadTemplateFile(templateName string) (tTemplate, error) {
	path, err := getTemplatesDirPath()
	if err != nil {
		return tTemplate{}, fmt.Errorf("%w: %w", ErrCannotReadTemplatesDir, err)
	}

	filePath := fmt.Sprintf("%s/%s.json", path, templateName)
	out, err := os.ReadFile(filePath)
	if err != nil {
		return tTemplate{}, fmt.Errorf("%w %s: %w", ErrCannotReadTemplateFile, filePath, err)
	}

	var t tTemplate
	err = json.Unmarshal(out, &t)
	if err != nil {
		return tTemplate{}, fmt.Errorf("%w: %w", ErrCannotUnmarshalData, err)
	}

	return t, nil
}

func loadTemplateFiles() ([]tTemplate, error) {
	var templates []tTemplate

	path, err := getTemplatesDirPath()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotReadTemplatesDir, err)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadTemplatesDir, path, err)
	}

	for _, e := range entries {
		filePath := fmt.Sprintf("%s/%s", path, e.Name())
		out, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("%w %s: %w", ErrCannotReadTemplateFile, filePath, err)
		}

		var t tTemplate
		err = json.Unmarshal(out, &t)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotUnmarshalData, err)
		}

		templates = append(templates, t)
	}

	return templates, nil
}

func saveTemplateFile(template tTemplate) error {
	hasTemplatesDir, err := hasTemplatesDir()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateTemplatesDir, err)
	}

	if !hasTemplatesDir {
		if err := createTemplatesDir(); err != nil {
			return fmt.Errorf("%w %s: %w", ErrCannotCreateTemplatesDir, templatesDir, err)
		}
	}

	j, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotMarshalData, err)
	}

	path, _ := getTemplatesDirPath()
	filePath := fmt.Sprintf("%s/%s.json", path, template.Name)
	err = os.WriteFile(filePath, j, 0644)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotSaveTemplate, filePath, err)
	}

	return nil
}

func deleteTemplateFile(templateName string) error {
	hasTemplatesDir, err := hasTemplatesDir()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotDeleteTemplate, templateName, err)
	}

	if !hasTemplatesDir {
		return fmt.Errorf("%w %s", ErrCannotDeleteTemplate, templateName)
	}

	path, _ := getTemplatesDirPath()
	filePath := fmt.Sprintf("%s/%s.json", path, templateName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("%w %s: %w", ErrCannotDeleteTemplate, filePath, err)
	}

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotDeleteTemplate, filePath, err)
	}

	return nil
}
