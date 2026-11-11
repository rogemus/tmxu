package config

import (
	"os"
)

// TODO: temp name remove 2 once everything is working
const sessionsFileName = "tmux-sessions2.json"

func (c Config) SaveSessions(data []byte) (string, error) {
	path, err := tmxuConfigDir(c.DirPath)
	if err != nil {
		return "", err
	}

	root, err := os.OpenRoot(path)
	if err != nil {
		return "", err
	}
	defer root.Close()

	return saveToFile(root, sessionsFileName, data)
}
