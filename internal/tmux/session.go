package tmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateSession   = errors.New("cannot create session")
	ErrCannotListSessions    = errors.New("cannot list sessions")
	ErrCannotAttachToSession = errors.New("cannot attach to session")
	ErrSessionExists         = errors.New("session already exists")
	ErrCannotValidateSession = errors.New("cannot validate session")
)

type Session struct {
	ID   string `json:"-"`
	Slug string `json:"-"`
	Name string `json:"session_name"`
}

func (s Session) GetName() string { return utils.Quote(s.Name) }

func NewSession(name string, dir string) error {
	out, err := exec.Command("tmux", buildNewSessionCmd(name, dir)...).CombinedOutput()
	if err != nil {
		return newSessionError(name, out, err)
	}

	return nil
}

func NewSessionFromTempl(name string, snapshot SnapshotData, dir string) error {
	out, err := exec.Command("tmux", buildNewSessionCmd(name, dir)...).CombinedOutput()
	if err != nil {
		return newSessionError(name, out, err)
	}

	sSlug := sessionSlug(name)
	windows, _ := snapshot.Windows[sSlug]
	return CreateWindows(name, windows, snapshot.Panes)
}

func ListSessions() ([]Session, error) {
	format := "{" + strings.Join(
		[]string{
			`"session_id": "#{session_id}"`,
			`"session_name": "#{session_name}"`,
			`"session_windows": #{session_windows}`,
		}, ",",
	) + "},"

	o, err := exec.Command("tmux", "list-sessions", "-F", format).Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	sess := "[" + strings.TrimSuffix(strings.TrimSpace(string(o)), ",") + "]"
	var sessions []Session
	err = json.Unmarshal([]byte(sess), &sessions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	return sessions, nil
}

func AttachToSession(name string) error {
	cmd := exec.Command("tmux", "attach", "-t", "="+name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotAttachToSession, name, err)
	}

	return nil
}

func CreateSessions(sessions []Session) error {
	if len(sessions) == 0 {
		return nil
	}

	cmd := buildNewSessionsCmd(sessions)
	_, err := exec.Command("tmux", cmd).Output()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateSession, err)
	}

	return nil
}

func HasSession(sessionName string) (bool, error) {
	output, err := exec.Command("tmux", "has-session", "-t", sessionName).Output()
	if err != nil {
		return false, fmt.Errorf("%w %s: %w", ErrCannotValidateSession, sessionName, err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return true, nil
	}

	return false, nil
}

func sessionSlug(sessionName string) string {
	return utils.Slug(sessionName)
}

func buildNewSessionsCmd(sessions []Session) string {
	cmds := make([]string, 0)

	for _, s := range sessions {
		cmd := strings.Join([]string{"new-session -d -s", s.GetName()}, " ")
		cmds = append(cmds, cmd)
	}

	return strings.Join(cmds, " ; ")
}

func buildNewSessionCmd(name, dir string) []string {
	argv := []string{"new-session", "-d", "-s", name}

	if dir != "" {
		argv = append(argv, "-c", dir)
	}

	return argv
}

// SessionExistsError reports that a session with the given name is already running.
// It unwraps to ErrSessionExists.
type SessionExistsError struct {
	Name string
}

func (e *SessionExistsError) Error() string {
	return fmt.Sprintf("%s: %s", ErrSessionExists, e.Name)
}

func (e *SessionExistsError) Unwrap() error { return ErrSessionExists }

func newSessionError(name string, out []byte, err error) error {
	msg := strings.TrimSpace(string(out))
	if strings.HasPrefix(msg, "duplicate session:") {
		return &SessionExistsError{Name: name}
	}

	return fmt.Errorf("%w %s: %s: %w", ErrCannotCreateSession, name, msg, err)
}
