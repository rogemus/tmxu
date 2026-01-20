package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configDir = "/.config/tmxu/"
const sessionFile = "tmux-sessions.json"

func saveFile(data []tSession) error {
	hasConfigDir, err := hasConfigDir()
	if err != nil {
		return fmt.Errorf("cannot check for config dir at path: ~%s", configDir)
	}

	if !hasConfigDir {
		if err := createConfigDir(); err != nil {
			return fmt.Errorf("cannot create config dir: ~%s", configDir)
		}
	}

	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal sassion data")
	}

	path, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("unable to get file path")
	}

	err = os.WriteFile(path, j, 0644)
	if err != nil {
		return fmt.Errorf("cannot save session file at path: %s", path)
	}

	return nil
}

func hasConfigDir() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("unable to get home dir")
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
		return fmt.Errorf("unable to get home dir")
	}

	path := filepath.Join(homeDir, configDir)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("unable to create tmxu config dir: %s", path)
	}

	return nil
}

func loadFile() ([]tSession, error) {
	var data []tSession

	path, err := getSessionFilePath()
	if err != nil {
		return nil, fmt.Errorf("unable to get file path")
	}

	out, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read tmux session file at path: %s", path)
	}

	err = json.Unmarshal(out, &data)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal sassion data")
	}

	return data, nil
}

func getSessionFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get home dir")
	}

	return filepath.Join(homeDir, configDir, sessionFile), nil
}

func loadTemplateFiles() ([]tTemplate, error) {
	// var templates []tTemplate

	// sessionToTemplate

	return nil, nil
}

func saveTemplateFile(template tTemplate) error {
	return nil
}

func deleteTemplateFile(templateName string) error {
	return nil
}
