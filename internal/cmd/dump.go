package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = Dump{}

var (
	ErrCannotParseDumpCmdFlags = fmt.Errorf("%w: cannot parse dump cmd flags", ErrUsage)
)

type Dump struct {
	cfg config.Config
	ios IOs
}

func NewDump(cfg config.Config, io IOs) Dump {
	return Dump{cfg, io}
}

var dumpInfo = Info{
	Command:   "dump",
	Aliases:   []string{"d"},
	Arg:       "",
	DescShort: "Print running tmux sessions, windows and panes",
	DescLong:  "Reads the live tmux state and prints every session with its windows and panes. Windows with several panes include their layout as percentages. Nothing is saved. With -json the output is indented JSON without tmux ids.",
	Flags: [][]string{
		{"json", "Print JSON instead of text"},
	},
	Examples: []string{
		"tmxu dump",
		"tmxu dump -json",
	},
}

func (c Dump) Info() Info { return dumpInfo }
func (c Dump) Run(args []string) error {
	var asJSON bool

	fs := flag.NewFlagSet("dump", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&asJSON, "json", false, "Print JSON instead of text")
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseDumpCmdFlags, err)
	}

	d, err := tmux.GetTmuxData()
	if err != nil {
		return err
	}

	if asJSON {
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
