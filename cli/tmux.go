package main

import (
	"errors"
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
	Order         int16  `json:"order"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	SessionName   string `json:"sessionName"`
	SessionWindow string `json:"sessionWindow"`
}

func newTPane(tmuxPane, sessionName, sessionWindow string) (tPane, error) {
	parts := strings.Split(tmuxPane, " ")

	order, err := strconv.Atoi(parts[0])
	if err != nil {
		return tPane{}, fmt.Errorf("unable to parse order for pane: %s", parts[1])
	}

	return tPane{
		Order:         int16(order),
		Name:          parts[1],
		Path:          parts[2],
		SessionWindow: sessionWindow,
		SessionName:   sessionName,
	}, nil
}

var errorSessionExists = errors.New("session exists")

func ListSessions() ([]string, error) {
	output, err := exec.Command("tmux", "list-sessions", "-F", "#{session_id} #{session_name}").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to list tmux sessions")
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil
}

func NewSession(session tSession) error {
	hs, _ := HasSession(session.Name)
	if hs == true {
		return errorSessionExists
	}

	err := exec.Command("tmux", "new-session", "-d", "-s", session.Name).Run()
	if err != nil {
		return fmt.Errorf("unable to create session: %s %s \n", session.Name, err.Error())
	}

	fmt.Printf("Session created: %s \n", session.Name)

	return nil
}

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
		return false, fmt.Errorf("unable to validate session: %s", sessionName)
	}

	if strings.TrimSpace(string(output)) == "" {
		return true, nil
	}

	return false, nil
}

func NewWindow(window tWindow) error {
	if window.Order == 1 {
		err := RenameWindow(window)

		if err != nil {
			return fmt.Errorf("unable to rename initial window in the session: %s \n", window.SessionName)
		}

		fmt.Printf("  Window renamed: %s in session: %s \n", window.Name, window.SessionName)
		return nil
	} else {
		err := exec.Command("tmux", "new-window", "-t", window.SessionWindow, "-n", window.Name).Run()
		if err != nil {
			return fmt.Errorf("unable to create window: %s \n", window.Name)
		}
	}

	err := SetWindowLayout(window)
	if err != nil {
		return fmt.Errorf("unable to set layout for window: %s \n", window.SessionWindow)
	}

	fmt.Printf(" Window created: %s \n", window.Name)
	return nil
}

func SetWindowLayout(window tWindow) error {
	err := exec.Command("tmux", "select-layout", "-t", window.SessionWindow, window.Layout).Run()
	if err != nil {
		return fmt.Errorf("unable to select layout for window: %s", window.SessionWindow)
	}

	return nil
}

func RenameWindow(window tWindow) error {
	err := exec.Command("tmux", "rename-window", "-t", window.SessionWindow, window.Name).Run()
	if err != nil {
		return fmt.Errorf("unable to rename window: %s", window.SessionWindow)
	}

	return nil
}

func ListPanes(sessionWindow string) ([]string, error) {
	output, err := exec.Command("tmux", "list-panes", "-t", sessionWindow, "-F", "#{pane_index} #{pane_title} #{pane_current_path}").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to list panes for window: %s \n", sessionWindow)
	}

	return strings.Split(
		strings.TrimSpace(string(output)), "\n",
	), nil

}

func NewPane(pane tPane) error {
	if pane.Order != 1 {
		err := exec.Command("tmux", "split-window", "-d", "-c", pane.Path, "-t", pane.SessionWindow).Run()
		if err != nil {
			return fmt.Errorf("unable to create pane: %s for window: %s \n", pane.Name, pane.SessionWindow)
		}
	}

	err := RenamePane(pane)
	if err != nil {
		return err
	}

	fmt.Printf("    Pane: %s created for window: %s \n", pane.Name, pane.SessionWindow)
	return nil
}

func RenamePane(pane tPane) error {
	targetPane := fmt.Sprintf("%s.%d", pane.SessionWindow, pane.Order)
	err := exec.Command("tmux", "select-pane", "-t", targetPane, "-T", pane.Name).Run()
	if err != nil {
		return fmt.Errorf("unable to rename pane: %s \n", targetPane)
	}

	return nil
}
