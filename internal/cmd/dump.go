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
	DescLong:  "Reads the live tmux state and prints every session with its windows and panes. Windows with several panes include their layout as percentages. Nothing is saved. With -json the output is indented JSON without tmux ids.",
	Examples: []string{
		"tmxu dump",
		"tmxu dump -session sessionName",
		"tmxu dump -json",
	},
}

func (c *Dump) Flags() *flag.FlagSet { return c.fs }
func (c *Dump) Info() Info           { return dumpInfo }
func (c *Dump) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseDumpCmdFlags, err)
	}

	if c.session != "" {
		d, err := tmux.SessionSnapshot(c.session)
		if err != nil {
			return err
		}

		return dumpSnapshot(c.json, d, c.ios.Out)
	}

	d, err := tmux.Snapshot()
	if err != nil {
		return err
	}

	return dumpSnapshot(c.json, d, c.ios.Out)
}

func dumpSnapshot(json bool, data tmux.SnapshotData, w io.Writer) error {
	if json {
		return dumpAsJSON(data, w)
	}

	return dumpAsString(data, w)
}

func dumpAsString(data tmux.SnapshotData, out io.Writer) error {
	_, err := fmt.Fprintf(out, "tmxu dump %s\n", data.String())
	return err
}

func dumpAsJSON(data tmux.SnapshotData, w io.Writer) error {
	out, err := data.JSON()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s\n", out)
	return err
}
