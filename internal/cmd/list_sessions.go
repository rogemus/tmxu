package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = ListSessions{}

var (
	ErrCannotParseListSessionCmdFlags = fmt.Errorf("%w: cannot parse list sessions cmd flags", ErrUsage)
	ErrCannotParseListSessionCmdJson  = fmt.Errorf("%w: cannot parse list sessions cmd json", ErrUsage)
)

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
	Flags: [][]string{
		{"json", "Print JSON instead of text"},
	},
	Examples: []string{
		"tmxu list-sessions",
		"tmxu list",
		"tmxu ls",
	},
}

func (c ListSessions) Info() Info { return listSessionsInfo }
func (c ListSessions) Run(args []string) error {
	var asJSON bool

	fs := flag.NewFlagSet("dump", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&asJSON, "json", false, "Print JSON instead of text")
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseDumpCmdFlags, err)
	}

	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	if asJSON {
		sDump := map[string][]tmux.Session{
			"sessions": sessions,
		}

		out, err := json.MarshalIndent(sDump, "", "  ")
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCannotParseListSessionCmdJson, err)
		}

		_, err = fmt.Fprintf(c.ios.Out, string(out))
		return err
	}

	w := tabwriter.NewWriter(c.ios.Out, 2, 0, 2, ' ', 0)
	for _, s := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%d\n", s.ID, s.Name, s.WindowsCount)
	}

	return w.Flush()
}
