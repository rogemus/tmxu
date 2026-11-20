package tmux

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rogemus/tmxu/internal/utils"
)

var (
	ErrCannotCreateSession   = errors.New("cannot create session")
	ErrCannotListSessions    = errors.New("cannot list sessions")
	ErrCannotAttachToSession = errors.New("cannot attach to session")
	ErrSessionExists         = errors.New("session already exists")
	ErrNoSession             = errors.New("session not found")
	ErrCannotValidateSession = errors.New("cannot validate session")
	ErrInvalidTemplate       = errors.New("invalid template")
)

// sessionPartsCount is the number of fields in one `list-sessions` row. Keep it
// equal to the length of `fields` in buildListSessionsFormat.
const sessionPartsCount = 1

type SessionData struct {
	Name    string       `json:"session_name"`
	Windows []WindowData `json:"session_windows,omitempty"`
}

// AsTemplate clears the session name and every pane path, so the session can
// be saved as a template and reused under any name.
func (d SessionData) AsTemplate() SessionData {
	d.Name = ""

	for wi := range d.Windows {
		w := &d.Windows[wi]
		for pi := range w.Panes {
			w.Panes[pi].AsTemplate()
		}
	}

	return d
}

func NewSession(name string, dir string) error {
	return createSession(name, dir)
}

func createSession(name string, dir string) error {
	out, err := exec.Command("tmux", buildNewSessionCmd(name, dir)...).CombinedOutput()
	if err != nil {
		return newSessionError(name, out, err)
	}

	return nil
}

func ListSessions() ([]SessionData, error) {
	o, err := exec.Command("tmux", "list-sessions", "-F", buildListSessionsFormat()).Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	sessions, err := parseSessionLines(o)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	return sessions, nil
}

func AttachToSession(name string, in io.Reader, out, errw io.Writer) error {
	cmd := exec.Command("tmux", "attach", "-t", "="+name)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errw

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrCannotAttachToSession, name, err)
	}

	return nil
}

func CreateSessions(sessions []SessionData) error {
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
	err := exec.Command("tmux", "has-session", "-t", "="+sessionName).Run()

	if err != nil {
		var ee *exec.ExitError

		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return false, nil
		}

		return false, fmt.Errorf("%w %s: %w", ErrCannotValidateSession, sessionName, err)
	}

	return true, nil
}

func parseSessionLines(output []byte) ([]SessionData, error) {
	sessions := make([]SessionData, 0)
	outputLines := strings.SplitSeq(strings.TrimSuffix(string(output), "\n"), "\n")

	for outputLine := range outputLines {
		parts := strings.SplitN(outputLine, lineSafeSeparator, sessionPartsCount)
		if len(parts) != sessionPartsCount {
			return nil, fmt.Errorf("%w: missing parts %d out of %d", ErrCannotParseLine, len(parts), sessionPartsCount)
		}

		session := SessionData{
			Name: parts[0],
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// buildListSessionsFormat returns the `list-sessions -F` format: one row per
// session, fields joined by lineSafeSeparator. Field order must match
// parseSessionLines. session_name stays last so SplitN keeps any stray
// separator in it.
func buildListSessionsFormat() string {
	fields :=
		[]string{
			`#{session_name}`,
		}

	return strings.Join(fields, lineSafeSeparator)
}

func buildNewSessionsCmd(sessions []SessionData) string {
	cmds := make([]string, 0)

	for _, s := range sessions {
		name := utils.Quote(s.Name)
		cmd := strings.Join([]string{"new-session -d -s", name}, " ")
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

func newSessionError(name string, out []byte, err error) error {
	msg := strings.TrimSpace(string(out))
	if strings.HasPrefix(msg, "duplicate session:") {
		return fmt.Errorf("%w: %s", ErrSessionExists, name)
	}

	return fmt.Errorf("%w %s: %s: %w", ErrCannotCreateSession, name, msg, err)
}
