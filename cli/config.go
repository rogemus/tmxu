package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configDir = "/.config/tmxu/"
const templatesDir = "templates"
const sessionFile = "tmux-sessions.json"

func saveSessionsFile(data []tSession) error {
	hasConfigDir, err := hasConfigDir()
	if err != nil {
		return fmt.Errorf("Cannot check for config dir at path: ~%s \n", configDir)
	}

	if !hasConfigDir {
		if err := createConfigDir(); err != nil {
			return fmt.Errorf("Cannot create config dir: ~%s \n", configDir)
		}
	}

	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("Cannot marshal sassion data \n")
	}

	path, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("Unable to get file path \n")
	}

	err = os.WriteFile(path, j, 0644)
	if err != nil {
		return fmt.Errorf("Cannot save session file at path: %s \n", path)
	}

	return nil
}

func hasConfigDir() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("Unable to get home dir \n")
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
		return fmt.Errorf("Unable to get home dir \n")
	}

	path := filepath.Join(homeDir, configDir)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Unable to create tmxu config dir: %s ]n", path)
	}

	return nil
}

func loadSessionsFile() ([]tSession, error) {
	var data []tSession

	path, err := getSessionFilePath()
	if err != nil {
		return nil, fmt.Errorf("Unable to get file path \n")
	}

	out, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read tmux session file at path: %s \n", path)
	}

	err = json.Unmarshal(out, &data)
	if err != nil {
		return nil, fmt.Errorf("Cannot marshal sassion data \n")
	}

	return data, nil
}

func getSessionFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Unable to get home dir \n")
	}

	return filepath.Join(homeDir, configDir, sessionFile), nil
}

func hasTemplatesDir() (bool, error) {
	_, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("Unable to get home dir \n")
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
		return fmt.Errorf("Unable to get home dir \n")
	}

	path, _ := getTemplatesDirPath()
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("Unable to create tmxu templates dir: %s \n", path)
	}

	return nil
}

func getTemplatesDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Unable to get templates dir \n")
	}

	return filepath.Join(homeDir, configDir, templatesDir), nil
}

func loadTemplateFile(templateName string) (tTemplate, error) {
	path, err := getTemplatesDirPath()
	if err != nil {
		return tTemplate{}, fmt.Errorf("Cannot read templates dir \n")
	}

	filePath := fmt.Sprintf("%s/%s.json", path, templateName)
	out, err := os.ReadFile(filePath)
	if err != nil {
		return tTemplate{}, fmt.Errorf("Unable to read tmux template file at path: %s \n", filePath)
	}

	var t tTemplate
	err = json.Unmarshal(out, &t)
	if err != nil {
		return tTemplate{}, fmt.Errorf("Cannot unmarshal template data \n")
	}

	return t, nil
}

func loadTemplateFiles() ([]tTemplate, error) {
	var templates []tTemplate

	path, err := getTemplatesDirPath()
	if err != nil {
		return nil, fmt.Errorf("Cannot read templates dir \n")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read templates dir \n")
	}

	for _, e := range entries {
		filePath := fmt.Sprintf("%s/%s", path, e.Name())
		out, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("Unable to read tmux template file at path: %s \n", filePath)
		}

		var t tTemplate
		err = json.Unmarshal(out, &t)
		if err != nil {
			return nil, fmt.Errorf("Cannot unmarshal template data \n")
		}

		templates = append(templates, t)
	}

	return templates, nil
}

func saveTemplateFile(template tTemplate) error {
	hasTemplatesDir, err := hasTemplatesDir()
	if err != nil {
		return fmt.Errorf("unable to create templates dir")
	}

	if !hasTemplatesDir {
		if err := createTemplatesDir(); err != nil {
			return fmt.Errorf("Cannot create template dir: ~%s \n", templatesDir)
		}
	}

	j, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("Cannot marshal template data")
	}

	path, _ := getTemplatesDirPath()
	filePath := fmt.Sprintf("%s/%s.json", path, template.Name)
	err = os.WriteFile(filePath, j, 0644)
	if err != nil {
		return fmt.Errorf("Cannot save template file at path: %s \n", filePath)
	}

	return nil
}

func deleteTemplateFile(templateName string) error {
	hasTemplatesDir, err := hasTemplatesDir()
	if err != nil || !hasTemplatesDir {
		return fmt.Errorf("Unable to delete template \n")
	}

	path, _ := getTemplatesDirPath()
	filePath := fmt.Sprintf("%s/%s.json", path, templateName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("Unable to delete template: %s \n", filePath)
	}

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("Unable to delete template: %s \n", filePath)
	}

	return nil
}
