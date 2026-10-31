package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateSessions = errors.New("cannot create session")
)

type Session struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewSession(l Line) Session {
	return Session{
		ID:   l.SessionID,
		Name: l.SessionName,
	}
}

func (s Session) StrName() string { return utils.Quote(s.Name) }

func (s Session) Eq(c utils.Comperable) bool {
	i, ok := c.(Session)
	return ok && s.ID == i.ID
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
