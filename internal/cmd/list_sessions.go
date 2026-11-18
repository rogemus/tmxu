package cmd

import (
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*ListSessions)(nil)

var (
	ErrCannotParseListSessionCmdFlags = fmt.Errorf("%w: cannot parse list sessions cmd flags", ErrUsage)
)

type ListSessions struct {
	ios IOs

	fs   *flag.FlagSet
	json bool
}

func NewListSessions(ios IOs) *ListSessions {
	cmd := &ListSessions{
		ios: ios,
		fs:  flag.NewFlagSet("list-sessions", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")

	return cmd
}

var listSessionsInfo = Info{
	Command:   "list-sessions",
	Aliases:   []string{"list", "ls"},
	Arg:       "",
	DescShort: "List all active sessions in tmux",
	DescLong:  "Displays all currently running tmux sessions with their names and window counts. With -json prints {\"sessions\":[...]} on one line.",
	Examples: []string{
		"tmxu list-sessions",
		"tmxu list",
		"tmxu ls",
		"tmxu ls -json",
	},
}

func (c *ListSessions) Info() Info           { return listSessionsInfo }
func (c *ListSessions) Flags() *flag.FlagSet { return c.fs }
func (c *ListSessions) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseListSessionCmdFlags, err)
	}

	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	if c.json {
		return printJSON(c.ios.Out, map[string][]tmux.Session{"sessions": sessions})
	}

	w := tabwriter.NewWriter(c.ios.Out, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Session Name\tWindows Count")
	for _, s := range sessions {
		fmt.Fprintf(w, "%s\t%d\n", s.Name, s.WindowsCount)
	}

	return w.Flush()
}
