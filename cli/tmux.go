package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCannotCreateSession   = errors.New("cannot create session")
	ErrCannotKillSession     = errors.New("cannot kill session")
	ErrCannotAttachSession   = errors.New("cannot attach to session")
	ErrCannotListSessions    = errors.New("cannot list sessions")
	ErrCannotCreatePane      = errors.New("cannot create pane")
	ErrCannotRenamePane      = errors.New("cannot rename pane")
	ErrCannotListPanes       = errors.New("cannot list panes")
	ErrCannotListWindows     = errors.New("cannot list windows for session")
	ErrCannotValidateSession = errors.New("cannot validate session")
	ErrCannotCreateWindow    = errors.New("cannot create window")
	ErrCannotRenameWindow    = errors.New("cannot rename window")
	ErrCannotSetLayout       = errors.New("cannot set window layout")
	ErrCannotParseWindow     = errors.New("cannot parse window")
	ErrCannotParsePane       = errors.New("cannot parse pane")
	ErrSessionExists         = errors.New("session exists")
	ErrInvalidTimestamp      = errors.New("invalid timestamp")
	ErrInvalidWindowsIndex   = errors.New("invalid windows index")
	ErrInvalidSessionIndex   = errors.New("invalid session index")
)

type tSession struct {
	Order   int16     `json:"order"`
	Name    string    `json:"name"`
	Windows []tWindow `json:"windows"`
}

type tSessionSimple struct {
	Created time.Duration
	Name    string
	Windows int
}

func newTSessionSimple(s string) (tSessionSimple, error) {
	parts := strings.SplitN(s, "\t", 3)
	// Parse Unix timestamp from tmux (seconds since epoch)
	createdTimestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return tSessionSimple{}, ErrInvalidTimestamp
	}

	createdTime := time.Unix(createdTimestamp, 0)
	duration := time.Since(createdTime)
	windows, err := strconv.Atoi(parts[1])
	if err != nil {
		return tSessionSimple{}, ErrInvalidWindowsIndex
	}

	return tSessionSimple{
		Created: duration,
		Name:    parts[2],
		Windows: windows,
	}, nil
}

func (s tSessionSimple) Title() string {
	return s.Name
}

func (s tSessionSimple) Desc() string {
	since := timeSince(s.Created)

	return fmt.Sprintf("%d win · %s", s.Windows, since)
}

type tTemplate = tSession

func newTSession(tmuxSession string, order int) (tSession, error) {
	parts := strings.SplitN(tmuxSession, "\t", 3)

	return tSession{
		Order: int16(order),
		Name:  parts[2],
	}, nil
}

type tWindow struct {
	Order         int16   `json:"order"`
	Name          string  `json:"name"`
	Layout        string  `json:"layout"`
	SessionName   string  `json:"sessionName"`
	SessionWindow string  `json:"sessionWindow"`
	Panes         []tPane `json:"panes"`
}

func newTWindow(tmuxWindow, sessionName string) (tWindow, error) {
	parts := strings.SplitN(tmuxWindow, "\t", 3)

	order, err := strconv.Atoi(parts[0])
	if err != nil {
		return tWindow{}, fmt.Errorf("%w %s: %w", ErrCannotParseWindow, parts[2], err)
	}

	return tWindow{
		Order:         int16(order),
		Name:          parts[2],
		Layout:        parts[1],
		SessionName:   sessionName,
		SessionWindow: fmt.Sprintf("%s:%s", sessionName, parts[0]),
	}, nil
}

type tPane struct {
	Order         int16  `json:"order"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	SessionName   string `json:"sessionName"`
	SessionWindow string `json:"sessionWindow"`
}

// paneName derives a stable pane name from its window. Pane titles are not
// used because programs (shells, Claude Code) overwrite them dynamically.
func paneName(windowName string, order int16, paneCount int) string {
	if paneCount <= 1 {
		return windowName
	}

	return fmt.Sprintf("%s_%d", windowName, order)
}

func newTPane(tmuxPane, sessionName, sessionWindow, windowName string, paneCount int) (tPane, error) {
	parts := strings.SplitN(tmuxPane, "\t", 3)
	if len(parts) < 3 {
		return tPane{}, fmt.Errorf("%w %s", ErrCannotParsePane, tmuxPane)
	}

	order, err := strconv.Atoi(parts[0])
	if err != nil {
		return tPane{}, fmt.Errorf("%w %s: %w", ErrCannotParsePane, parts[0], err)
	}

	return tPane{
		Order:         int16(order),
		Name:          paneName(windowName, int16(order), paneCount),
		Path:          parts[1],
		SessionWindow: sessionWindow,
		SessionName:   sessionName,
	}, nil
}

func ListSessions() ([]string, error) {
	output, err := exec.Command("tmux", "list-sessions", "-F", "#{session_created}\t#{session_windows}\t#{session_name}").Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotListSessions, err)
	}

	ss := strings.Split(
		strings.TrimSpace(string(output)), "\n",
	)

	sort.Slice(ss, func(i, j int) bool {
		partsI := strings.SplitN(ss[i], "\t", 3)
		partsJ := strings.SplitN(ss[j], "\t", 3)
		tsI, _ := strconv.ParseInt(partsI[0], 10, 64)
		tsJ, _ := strconv.ParseInt(partsJ[0], 10, 64)
		return tsI < tsJ
	})

	return ss, nil
}

func NewSession(session tSession, force bool) error {
	hs, _ := HasSession(session.Name)
	if hs && !force {
		return fmt.Errorf("%w %s", ErrSessionExists, session.Name)
	} else if hs && force {
		err := exec.Command("tmux", "kill-session", "-t", session.Name).Run()
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrCannotKillSession, session.Name, err)
		}
	}

	startingDir := ""
	if len(session.Windows) > 0 {
		startingDir = session.Windows[0].Panes[0].Path
	}

	args := append([]string{"new-session"}, startDirArgs(startingDir)...)
	err := exec.Command("tmux", append(args, "-d", "-s", session.Name)...).Run()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotCreateSession, session.Name, err)
	}

	return nil
}

func AttachToSession(sessionName string) error {
	cmd := exec.Command("tmux", "attach", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotAttachSession, sessionName, err)
	}

	return nil
}

func ListWindows(sessionName string) ([]string, error) {
	output, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}\t#{window_layout}\t#{window_name}").Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotListWindows, sessionName, err)
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil
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

// startDirArgs returns the tmux `-c` args for path, or nothing when path is not
// an existing directory so tmux inherits the session's working directory.
func startDirArgs(path string) []string {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return nil
	}

	return []string{"-c", path}
}

// NewWindow renames the window created with the session when first is true,
// otherwise it creates a new one. The first window is targeted by session name
// because its saved index may differ from the current base-index.
func NewWindow(window tWindow, first bool) error {
	if first {
		initial := window
		initial.SessionWindow = window.SessionName
		err := RenameWindow(initial)

		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrCannotRenameWindow, window.SessionName, err)
		}

		return nil
	} else {
		args := append(startDirArgs(window.Panes[0].Path), "-t", window.SessionWindow, "-n", window.Name)
		err := exec.Command("tmux", append([]string{"new-window"}, args...)...).Run()
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrCannotCreateWindow, window.Name, err)
		}
	}

	return nil
}

func SetWindowLayout(window tWindow) error {
	err := exec.Command("tmux", "select-layout", "-t", window.SessionWindow, window.Layout).Run()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotSetLayout, window.SessionWindow, err)
	}

	return nil
}

func RenameWindow(window tWindow) error {
	err := exec.Command("tmux", "rename-window", "-t", window.SessionWindow, window.Name).Run()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotRenameWindow, window.SessionWindow, err)
	}

	return nil
}

func ListPanes(sessionWindow string) ([]string, error) {
	output, err := exec.Command("tmux", "list-panes", "-t", sessionWindow, "-F", "#{pane_index}\t#{pane_current_path}\t#{pane_title}").Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotListPanes, sessionWindow, err)
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil

}

func NewPane(pane tPane) error {
	if pane.Order != 1 {
		args := append([]string{"split-window", "-d"}, startDirArgs(pane.Path)...)
		err := exec.Command("tmux", append(args, "-t", pane.SessionWindow)...).Run()
		if err != nil {
			return fmt.Errorf("%w %s|%s: %w", ErrCannotCreatePane, pane.SessionWindow, pane.Name, err)
		}
	}

	err := RenamePane(pane)
	if err != nil {
		return err
	}

	return nil
}

func RenamePane(pane tPane) error {
	targetPane := fmt.Sprintf("%s.%d", pane.SessionWindow, pane.Order)
	err := exec.Command("tmux", "select-pane", "-t", targetPane, "-T", pane.Name).Run()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCannotRenamePane, targetPane, err)
	}

	return nil
}
