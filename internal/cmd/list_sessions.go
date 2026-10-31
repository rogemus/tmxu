package cmd

import (
	"errors"
	"fmt"

	"github.com/rogemus/tmxu/internal/tmux"
)

var (
	ErrCannotListSessionCmd = errors.New("cannot list sessions in cmd")
)

var _ Cmd = ListSessionsCmd{}

// ListSessionsCmd prints the running tmux sessions.
type ListSessionsCmd struct{}

// NewListSessionsCmd builds a ListSessionsCmd. It needs no config.
func NewListSessionsCmd() ListSessionsCmd {
	return ListSessionsCmd{}
}

func (c ListSessionsCmd) Command() string { return "list-sessions" }

func (c ListSessionsCmd) Aliases() []string { return []string{"list", "ls"} }

func (c ListSessionsCmd) Arg() string { return "" }

func (c ListSessionsCmd) DescShort() string { return "List all active sessions in tmux" }

func (c ListSessionsCmd) DescLong() string {
	return "Displays all currently running tmux sessions with their IDs and names."
}

func (c ListSessionsCmd) Flags() [][]string { return nil }

func (c ListSessionsCmd) Examples() []string {
	return []string{
		"tmxu list-sessions",
		"tmxu list",
		"tmxu ls",
	}
}

func (c ListSessionsCmd) Run() error {
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
