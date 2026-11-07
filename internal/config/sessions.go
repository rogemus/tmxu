package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrCannotSaveSessionFile = errors.New("cannot save session file")
)

func (c Config) SaveSessions(data []byte) error {
	path := c.SessionsFileName()

	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotSaveSessionFile, path, err)
	}
	return nil
}
