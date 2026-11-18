package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = (*SaveSessions)(nil)

var (
	ErrCannotParseSaveSessionsCmdFlags = fmt.Errorf("%w: cannot parse save sessions cmd flags", ErrUsage)
)

type SaveSessions struct {
	ios IOs
	cfg config.Config

	fs  *flag.FlagSet
	yes bool
}

// NewSaveSessions builds a SaveSessionsCmd from the app config.
func NewSaveSessions(ios IOs, cfg config.Config) *SaveSessions {
	cmd := &SaveSessions{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("save-sessions", flag.ContinueOnError),
	}

	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.yes, "yes", false, "Skip the confirmation prompt")

	return cmd
}

var saveSessionsInfo = Info{
	Command:   "save-sessions",
	Aliases:   []string{"save", "s"},
	Arg:       "",
	DescShort: "Save tmux sessions",
	DescLong:  "Captures all running tmux sessions including windows, panes, and layouts. Asks for confirmation first; -yes skips the prompt. Saves to ~/.config/tmxu/tmux-sessions.json.",
	Examples: []string{
		"tmxu save-sessions",
		"tmxu save",
		"tmxu s",
		"tmxu save -yes",
	},
}

func (c *SaveSessions) Flags() *flag.FlagSet { return c.fs }
func (c *SaveSessions) Info() Info           { return saveSessionsInfo }
func (c *SaveSessions) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseSaveSessionsCmdFlags, err)
	}

	if !c.yes && !utils.Confirm("Save all tmux sessions?", c.ios.In, c.ios.Err) {
		fmt.Fprintln(c.ios.Err, "Aborted")
		return nil
	}

	s, err := tmux.Snapshot()
	if err != nil {
		return err
	}

	out, err := s.JSON()
	if err != nil {
		return err
	}

	fPath, err := c.cfg.SaveSessions(out)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Tmux sessions saved at %s\n", fPath)
	return err
}
