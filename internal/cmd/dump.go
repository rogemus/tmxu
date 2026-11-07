package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = Dump{}

var (
	ErrCannotParseDumpCmdFlags = fmt.Errorf("%w: cannot parse dump cmd flags", ErrUsage)
)

type Dump struct {
	ios IOs

	fs   *flag.FlagSet
	json bool
}

func NewDump(ios IOs) Dump {
	cmd := Dump{ios: ios}

	cmd.fs = flag.NewFlagSet("dump", flag.ContinueOnError)
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")

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
		"tmxu dump -json",
	},
}

func (c Dump) Flags() *flag.FlagSet { return c.fs }
func (c Dump) Info() Info           { return dumpInfo }
func (c Dump) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseDumpCmdFlags, err)
	}

	d, err := tmux.Snapshot()
	if err != nil {
		return err
	}

	if c.json {
		out, err := d.JSON()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(c.ios.Out, "%s\n", out)
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "tmxu dump %s\n", d.String())
	return err
}
