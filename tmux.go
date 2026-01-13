package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type tSession struct {
	Order   int16     `json:"order"`
	Name    string    `json:"name"`
	Windows []tWindow `json:"windows"`
}

func newTSession(tmuxSession string) (tSession, error) {
	parts := strings.Split(tmuxSession, " ")

	order, err := strconv.Atoi(strings.TrimPrefix(parts[0], "$"))
	if err != nil {
		return tSession{}, fmt.Errorf("unable to convert session id to order for session: %s", parts[1])
	}

	return tSession{
		Order: int16(order),
		Name:  parts[1],
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
	parts := strings.Split(tmuxWindow, " ")

	order, err := strconv.Atoi(parts[0])
	if err != nil {
		return tWindow{}, fmt.Errorf("unable to parse order for window: %s", parts[1])
	}

	return tWindow{
		Order:         int16(order),
		Name:          parts[1],
		Layout:        parts[2],
		SessionName:   sessionName,
		SessionWindow: fmt.Sprintf("%s:%s", sessionName, parts[0]),
	}, nil
}

type tPane struct {
	Order int16  `json:"order"`
	Name  string `json:"name"`
	Path  string `json:"path"`
}

func newTPane(tmuxPane string) (tPane, error) {
	parts := strings.Split(tmuxPane, " ")

	order, err := strconv.Atoi(parts[0])
	if err != nil {
		return tPane{}, fmt.Errorf("unable to parse order for pane: %s", parts[1])
	}

	return tPane{
		Order: int16(order),
		Name:  parts[1],
		Path:  parts[2],
	}, nil
}

func ListSessions() ([]string, error) {
	output, err := exec.Command("tmux", "list-sessions", "-F", "#{session_id} #{session_name}").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to list tmux sessions")
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil
}

func NewSession() {}

func AttachToSession(sessionName string) error {
	err := exec.Command("tmux", "attach", "-t", sessionName).Run()
	if err != nil {
		return fmt.Errorf("unable to attach to tmux session: %s", sessionName)
	}

	return nil
}

func ListWindows(sessionName string) ([]string, error) {
	output, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index} #{window_name} #{window_layout}").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to list tmux windows for session: %s", sessionName)
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil
}

func HasSession(sessionName string) (bool, error) {
	output, err := exec.Command("tmux", "has-session", "-t", sessionName).Output()
	if err != nil {
		return false, fmt.Errorf("unable to validate session: %s \n", sessionName)
	}

	if strings.TrimSpace(string(output)) == "" {
		return true, nil
	}

	return false, nil
}

func NewWindow()       {}
func SetWindowLayout() {}
func RenameWindow()    {}

func ListPanes(sessionWindow string) ([]string, error) {
	output, err := exec.Command("tmux", "list-panes", "-t", sessionWindow, "-F", "#{pane_index} #{pane_title} #{pane_current_path}").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to list panes for window: %s \n", sessionWindow)
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil

}
func NewPane()    {}
func RenamePane() {}
