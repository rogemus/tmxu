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

type ListSessions struct{}

func NewListSessions() ListSessions {
	return ListSessions{}
}

var listSessionsInfo = Info{
	Command:   "list-sessions",
	Aliases:   []string{"list", "ls"},
	Arg:       "",
	DescShort: "List all active sessions in tmux",
	DescLong:  "Displays all currently running tmux sessions with their IDs and names.",
	Flags:     nil,
	Examples: []string{
		"tmxu list-sessions",
		"tmxu list",
		"tmxu ls",
	},
}

func (c ListSessions) Info() Info { return listSessionsInfo }
func (c ListSessions) Run(args []string) error {
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
