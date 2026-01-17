package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const sessionFile = "/.config/tmux/tmux-session.json"

func saveFile(data []tSession) error {
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

	return filepath.Join(homeDir, sessionFile), nil
}
