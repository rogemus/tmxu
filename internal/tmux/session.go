package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrCannotCreateSessions = errors.New("cannot create session")
)

type Session struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	FirstWindowName string `json:"firstWindowName"`
	FirstWindowPath string `json:"firstWindowPath"`
}

func NewSession(l TmuxLine) Session {
	return Session{
		ID:              l.SessionID,
		Name:            l.SessionName,
		FirstWindowName: l.WindowName,
		FirstWindowPath: l.PanePath,
	}
}

func (s Session) Eq(c Comperable) bool {
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
		cmd := strings.Join([]string{"new-session -d -s", s.Name, "-n", s.FirstWindowName, "-c", s.FirstWindowPath}, " ")
		cmds = append(cmds, cmd)
	}

	return strings.Join(cmds, " ; ")
}
