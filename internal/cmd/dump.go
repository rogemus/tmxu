package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*Dump)(nil)

var (
	ErrCannotParseDumpCmdFlags = fmt.Errorf("%w: cannot parse dump cmd flags", ErrUsage)
)

type Dump struct {
	ios IOs

	fs      *flag.FlagSet
	json    bool
	session string
}

func NewDump(ios IOs) *Dump {
	cmd := &Dump{
		ios: ios,
		fs:  flag.NewFlagSet("dump", flag.ContinueOnError),
	}

	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")
	cmd.fs.StringVar(&cmd.session, "session", "", "Dump single session")

	return cmd
}

var dumpInfo = Info{
	Command:   "dump",
	Aliases:   []string{"d"},
	Arg:       "",
	DescShort: "Print running tmux sessions, windows and panes",
	DescLong:  "Reads the live tmux state and prints it as a tree: each session with its windows, and each window with its panes (pane name and start directory). Nothing is saved. Use -session to dump a single session. With -json the output is the compact JSON that save-sessions writes, including window layouts.",
	Examples: []string{
		"tmxu dump",
		"tmxu dump -json",
		"tmxu dump -session sessionName",
		"tmxu dump -session sessionName -json",
	},
}

func (c *Dump) Flags() *flag.FlagSet { return c.fs }
func (c *Dump) Info() Info           { return dumpInfo }
func (c *Dump) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseDumpCmdFlags, err)
	}

	d, err := tmux.Snapshot(c.session)
	if err != nil {
		return err
	}

	return dumpSnapshot(c.json, d, c.ios.Out)
}
