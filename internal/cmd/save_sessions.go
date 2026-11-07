package cmd

import (
	"flag"
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = (*SaveSessions)(nil)

type SaveSessions struct {
	ios IOs
	cfg config.Config
}

// NewSaveSessions builds a SaveSessionsCmd from the app config.
func NewSaveSessions(ios IOs, cfg config.Config) *SaveSessions {
	return &SaveSessions{ios: ios, cfg: cfg}
}

var saveSessionsInfo = Info{
	Command:   "save-sessions",
	Aliases:   []string{"save", "s"},
	Arg:       "",
	DescShort: "Save tmux sessions",
	DescLong:  "Captures all running tmux sessions including windows, panes, and layouts. Saves to ~/.config/tmxu/tmux-sessions.json.",
	Examples: []string{
		"tmxu save-sessions",
		"tmxu save",
		"tmux s",
	},
}

func (c *SaveSessions) Flags() *flag.FlagSet { return nil }
func (c *SaveSessions) Info() Info           { return saveSessionsInfo }
func (c *SaveSessions) Run(args []string) error {
	if !utils.Confirm("Save all tmux sessions?", c.ios.In, c.ios.Err) {
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

	fPath := c.cfg.SessionsFileName()
	err = c.cfg.SaveSessions(out)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.ios.Out, "Tmux sessions saved at %s\n", fPath)
	return nil
}
