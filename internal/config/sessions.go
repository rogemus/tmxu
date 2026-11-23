package config

import (
	"os"
	"path/filepath"
)

// TODO: temp name remove 2 once everything is working
const sessionsFileName = "tmux-sessions2.json"

func (c Config) SessionsFile() string {
	return filepath.Join(c.AppDirPath, sessionsFileName)
}

func (c Config) SaveSessions(data []byte) (string, error) {
	err := ensureTmxuConfigDir(c.AppDirPath)
	if err != nil {
		return "", err
	}

	root, err := os.OpenRoot(c.AppDirPath)
	if err != nil {
		return "", err
	}
	defer root.Close()

	return saveToFile(root, sessionsFileName, data)
}

func (c Config) ReadSessionsFile() ([]byte, error) {
	root, err := os.OpenRoot(c.AppDirPath)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	return readFile(root, sessionsFileName)
}
