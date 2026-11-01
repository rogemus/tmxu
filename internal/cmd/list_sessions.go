package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = ListSessions{}

type ListSessions struct {
	cfg config.Config
	ios IOs
}

func NewListSessions(cfg config.Config, ios IOs) ListSessions {
	return ListSessions{cfg, ios}
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
		return err
	}

	w := tabwriter.NewWriter(c.ios.Out, 2, 0, 2, ' ', 0)
	for _, s := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%d\n", s.ID, s.Name, s.WindowsCount)
	}

	return w.Flush()
}
