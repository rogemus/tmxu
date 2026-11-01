package cmd

import (
	"errors"
	"fmt"

	"github.com/rogemus/tmxu/internal/tmux"
)

var (
	ErrCannotListSessionCmd = errors.New("cannot list sessions in cmd")
)

var _ Cmd = ListSessions{}

// ListSessions prints the running tmux sessions.
type ListSessions struct{}

// NewListSessions builds a ListSessionsCmd. It needs no config.
func NewListSessions() ListSessions {
	return ListSessions{}
}

func (c ListSessions) Command() string { return "list-sessions" }

func (c ListSessions) Aliases() []string { return []string{"list", "ls"} }

func (c ListSessions) Arg() string { return "" }

func (c ListSessions) DescShort() string { return "List all active sessions in tmux" }

func (c ListSessions) DescLong() string {
	return "Displays all currently running tmux sessions with their IDs and names."
}

func (c ListSessions) Flags() [][]string { return nil }

func (c ListSessions) Examples() []string {
	return []string{
		"tmxu list-sessions",
		"tmxu list",
		"tmxu ls",
	}
}

func (c ListSessions) Run() error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotListSessionCmd, err)
	}

	fmt.Println("Available sessions")
	for _, s := range sessions {
		fmt.Printf(" %s %-20s [win:%d]\n ", s.ID, s.Name, s.WindowsCount)
	}

	return nil
}
