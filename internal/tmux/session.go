package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateSessions = errors.New("cannot create session")
	ErrCannotListSessions   = errors.New("cannot list sessions")
)

type Session struct {
	ID           string `json:"session_id"`
	Name         string `json:"session_name"`
	WindowsCount int    `json:"session_windows"`
}

func NewSession(l Line) Session {
	return Session{
		ID:           l.SessionID,
		Name:         l.SessionName,
		WindowsCount: 0,
	}
}

func (s Session) StrName() string { return utils.Quote(s.Name) }

func (s Session) Eq(c utils.Comperable) bool {
	i, ok := c.(Session)
	return ok && s.ID == i.ID
}

func ListSessions() ([]Session, error) {
	format := "{" + strings.Join(
		[]string{
			`"session_id": "#{session_id}"`,
			`"session_name": "#{session_name}"`,
			`"session_windows": session_windows"`,
		}, ",",
	) + "},"

	o, err := exec.Command("tmux", "list-sessions", "-F", format).Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	var sessions []Session
	err = json.Unmarshal(o, &sessions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	return sessions, nil
}

func CreateSessions(sessions []Session) error {
	if len(sessions) == 0 {
		return nil
	}

	cmd := buildNewSessionsCmd(sessions)
	_, err := exec.Command("tmux", cmd).Output()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateSessions, err)
	}

	return nil
}

func buildNewSessionsCmd(sessions []Session) string {
	cmds := make([]string, 0)

	for _, s := range sessions {
		cmd := strings.Join([]string{"new-session -d -s", s.StrName()}, " ")
		cmds = append(cmds, cmd)
	}

	return strings.Join(cmds, " ; ")
}
