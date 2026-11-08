package config

import "path/filepath"

// TODO: temp name remove 2 once everything is working
const sessionsFileName = "tmux-sessions2.json"

func (c Config) SaveSessions(data []byte) (string, error) {
	path, err := tmxuConfigDir(c.DirPath)
	if err != nil {
		return "", err
	}

	sessionsFile := sessionFile(path)
	return saveToFile(sessionsFile, data)
}

func sessionFile(path string) string {
	return filepath.Join(path, sessionsFileName)
}
